package to

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-nfs/nfsv3/nfs"
	"github.com/go-nfs/nfsv3/nfs/rpc"
	"github.com/go-nfs/nfsv3/nfs/util"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

type flagsDir struct {
	inputPath      string
	nfsHost        string
	nfsMountFolder string
}

// ToServer function provides functionaltiy to transfer files or folders from local filesystem to NFS server.
func ToServer() *cli.Command {
	var nfsOpts flagsDir
	return &cli.Command{
		Name: "to",
		// Aliases: []string{"d"},
		Usage: "The 'to' command is used to copy files or folders from the local machine to NFS server.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &nfsOpts.inputPath,
				Name:        "input",
				Aliases:     []string{"i"},
				Usage:       "Input refers to the specific location of a folder or file that will be transferred.",
			},

			&cli.StringFlag{
				Destination: &nfsOpts.nfsHost,
				Name:        "host",
				Aliases:     []string{"t"},
				Usage:       "The IP address or DNS of the NFS server specifies the IP or hostname that can be used to access the NFS server.",
			},
			&cli.StringFlag{
				Destination: &nfsOpts.nfsMountFolder,
				Name:        "nfspath",
				Aliases:     []string{"p"},
				Usage:       "The NFS path denotes the destination directory on the NFS server where files will be copied to.",
			},
		},
		Action: func(ctx *cli.Context) error {
			basePath := filepath.Dir(nfsOpts.inputPath)
			mount, err := nfs.DialMount(nfsOpts.nfsHost, false)
			if err != nil {
				log.Fatalf("unable to dial MOUNT service: %v", err)
			}
			defer mount.Close()

			hostNameLocal, _ := os.Hostname()

			auth := rpc.NewAuthUnix(hostNameLocal, 0, 0)

			nfs, err := mount.Mount(nfsOpts.nfsMountFolder, auth.Auth())
			if err != nil {
				log.Fatalf("unable to mount volume: %v", err)
			}
			defer nfs.Close()

			if err = mount.Unmount(); err != nil {
				log.Fatalf("unable to unmount target: %v", err)
			}

			mount.Close()

			folders, files, err := getFoldersAndFiles(nfsOpts.inputPath, "")
			if err != nil {
				log.Fatal(err)
			}
			for _, v := range folders {
				_, err = nfs.Mkdir(v, os.ModePerm)
				// skip file exist error
				if err == os.ErrExist {
					err = nil
				}

			}
			for _, targetfile := range files {
				sf := filepath.Join(basePath, targetfile)
				// Copy file to destination
				if err = transferFile(nfs, sf, targetfile); err != nil {
					log.Fatalf("fail")

				}
			}
			return nil
		},
	}
}

func checkMark() func() {
	return func() {
		fmt.Printf("%s ✔ %s\n", "\033[32m", "\033[0m")
	}
}

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func getFoldersAndFiles(path string, basePath string) ([]string, []string, error) {
	var folders []string
	var files []string

	// Check if the path is a directory
	isDir, err := isDirectory(path)
	if err != nil {
		return nil, nil, err
	}

	// Check if the path is a file
	if !isDir {
		filePath := filepath.Join(basePath, filepath.Base(path))
		files = append(files, filePath)
		return folders, files, nil
	}

	// Extract the folder name from the path
	folderName := filepath.Base(path)

	// Append the folder path to the folders slice
	folders = append(folders, filepath.Join(basePath, folderName))

	// Iterate over the directory contents
	contents, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, nil, err
	}

	for _, contentPath := range contents {
		// Check if the content is a directory
		isDir, err := isDirectory(contentPath)
		if err != nil {
			log.Println(err)
			continue
		}

		if isDir {
			subfolderFolders, subfolderFiles, err := getFoldersAndFiles(contentPath, filepath.Join(basePath, folderName))
			if err != nil {
				log.Println(err)
				continue
			}

			folders = append(folders, subfolderFolders...)
			files = append(files, subfolderFiles...)
		} else {
			filePath := filepath.Join(basePath, folderName, filepath.Base(contentPath))

			files = append(files, filePath)
		}
	}

	return folders, files, nil
}

func transferFile(nfs *nfs.Target, srcfile string, targetfile string) error {

	sourceFile, err := os.Open(srcfile)
	if err != nil {
		util.Errorf("error opening source file: %s", err.Error())
		return err
	}

	// Calculate the ShaSum
	h := sha256.New()
	t := io.TeeReader(sourceFile, h)
	stat, _ := sourceFile.Stat()
	size := stat.Size()

	defer sourceFile.Close()

	// Customize the progress bar theme
	theme := progressbar.Theme{
		Saucer:        "[yellow]▖[reset][cyan]",
		SaucerPadding: " ",
	}

	progress := progressbar.NewOptions64(
		size,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(theme),
		progressbar.OptionSetDescription("Copying"+" "+"[green]"+srcfile+"[reset]"),
		progressbar.OptionSetWidth(25),
		progressbar.OptionShowBytes(true),
		progressbar.OptionOnCompletion(checkMark()),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSpinnerType(14),
	)

	wr, err := nfs.OpenFile(targetfile, os.ModePerm)
	if err != nil {
		util.Errorf("error opening target file: %s", err.Error())
		return err
	}
	defer wr.Close()

	// Copy files with progress size
	n, err := io.CopyN(wr, io.TeeReader(t, progress), int64(size))
	if err != nil {
		util.Errorf("error copying: n=%d, %s", n, err.Error())
		return err
	}
	expectedSum := h.Sum(nil)

	// Get the file we wrote and calculate the sum
	rdr, err := nfs.Open(targetfile)
	if err != nil {
		util.Errorf("error opening target file for verification: %v", err)
		return err
	}
	defer rdr.Close()

	h = sha256.New()
	t = io.TeeReader(rdr, h)

	_, err = io.Copy(io.Discard, t) // Discard the content since we only need the sum
	if err != nil {
		util.Errorf("error reading target file for verification: %v", err)
		return err
	}
	actualSum := h.Sum(nil)

	if !bytes.Equal(actualSum, expectedSum) {
		log.Fatalf("[Verification Error] Actual SHA=%x Expected SHA=%s", actualSum, expectedSum)
	}

	progress.Finish()
	return nil
}
