package to

import (
	"bytes"
	"crypto/sha256"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-nfs/nfsv3/nfs"
	"github.com/go-nfs/nfsv3/nfs/rpc"
	"github.com/kha7iq/ncp/internal/helper"
	"github.com/urfave/cli/v2"
)

type nfsConfg struct {
	inputPath      string
	nfsHost        string
	nfsMountFolder string
}

// ToServer function provides functionaltiy to transfer files or folders from local filesystem to NFS server.
func ToServer() *cli.Command {
	var nc nfsConfg
	return &cli.Command{
		Name:      "to",
		Usage:     "The 'to' command is used to copy files or folders from the local machine to NFS server.",
		UsageText: "ncp to --host 192.168.0.80 --nfspath data --input src",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &nc.inputPath,
				Name:        "input",
				Required:    true,
				Aliases:     []string{"i"},
				Usage:       "Input refers to the specific location of a folder or file that will be transferred.",
			},

			&cli.StringFlag{
				Destination: &nc.nfsHost,
				Name:        "host",
				Aliases:     []string{"t"},
				Required:    true,
				Usage:       "IP address or DNS of the NFS server specifies the IP or hostname that can be used to access the NFS server.",
			},
			&cli.StringFlag{
				Destination: &nc.nfsMountFolder,
				Required:    true,
				Name:        "nfspath",
				Aliases:     []string{"p"},
				Usage:       "NFS path denotes the destination directory on the NFS server where files will be copied to.",
			},
		},
		Action: func(ctx *cli.Context) error {
			truncate := ctx.Bool("turncate")
			u := ctx.Int("uid")
			g := ctx.Int("gid")
			uid, gid := helper.CheckUID(u, g)

			basePath := filepath.Dir(nc.inputPath)
			mount, err := nfs.DialMount(nc.nfsHost, false)
			if err != nil {
				log.Fatalf("unable to dial MOUNT service: %v", err)
			}
			defer mount.Close()

			hostNameLocal, _ := os.Hostname()

			auth := rpc.NewAuthUnix(hostNameLocal, uid, gid)

			nfs, err := mount.Mount(nc.nfsMountFolder, auth.Auth())
			if err != nil {
				log.Fatalf("unable to mount volume: %v", err)
			}
			defer nfs.Close()
			if err = mount.Unmount(); err != nil {
				log.Fatalf("nable to unmount target: %v", err)
			}
			mount.Close()

			folders, files, err := getFoldersAndFiles(nc.inputPath, "")
			if err != nil {
				log.Fatalf("unable to get list of files and folders %v", err)
			}
			for _, v := range folders {
				_, err = nfs.Mkdir(v, os.ModePerm)
				// skip file exist error
				if err == os.ErrExist {
					err = nil
				}
				if err != nil {
					return err // But return all other errors
				}
			}
			for _, targetfile := range files {
				sf := filepath.Join(basePath, targetfile)
				// Copy file to destination
				if err = transferFile(nfs, sf, targetfile, truncate); err != nil {
					log.Fatalf("fail to transfer files %v", err)
				}
			}
			return nil
		},
	}
}

// listFileAndFolders take a directory path and returns a slice containng files and another containing folders
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
			log.Fatalf("can not check dir/file attributes %v", err)
			continue
		}
		if isDir {
			subfolderFolders, subfolderFiles, err := getFoldersAndFiles(contentPath, filepath.Join(basePath, folderName))
			if err != nil {
				log.Fatalf("unable to get list %v", err)
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

// transferFile will take a source and target file path along with *nfs.Targe to transfer file
func transferFile(nfs *nfs.Target, srcfile string, targetfile string, turnication bool) error {
	var filePath string
	sourceFile, err := os.Open(srcfile)
	if err != nil {
		log.Fatalf("error opening source file: %s", err.Error())
	}

	// Calculate the ShaSum
	h := sha256.New()
	t := io.TeeReader(sourceFile, h)
	stat, _ := sourceFile.Stat()
	size := stat.Size()

	defer sourceFile.Close()

	if !turnication {
		filePath = srcfile
	} else {
		filePath = helper.TruncateFileName(srcfile)
	}

	progress := helper.ProgressBar(size, filePath, helper.CheckMark())

	wr, err := nfs.OpenFile(targetfile, os.ModePerm)
	if err != nil {
		log.Fatalf("error opening target file: %s", err.Error())
		return err
	}
	defer wr.Close()

	// Copy files with progress size
	n, err := io.CopyN(wr, io.TeeReader(t, progress), size)
	if err != nil {
		log.Fatalf("error copying: n=%d, %s", n, err.Error())
		return err
	}
	expectedSum := h.Sum(nil)

	// Get the file we wrote and calculate the sum
	rdr, err := nfs.Open(targetfile)
	if err != nil {
		log.Fatalf("error opening target file for verification: %v", err)
		return err
	}
	defer rdr.Close()

	h = sha256.New()
	t = io.TeeReader(rdr, h)

	_, err = io.Copy(io.Discard, t) // Discard the content since we only need the sum
	if err != nil {
		log.Fatalf("error reading target file for verification: %v", err)
		return err
	}
	actualSum := h.Sum(nil)

	if !bytes.Equal(actualSum, expectedSum) {
		log.Fatalf("[Verification Error] Actual SHA=%x Expected SHA=%s", actualSum, expectedSum)
	}

	progress.Finish()
	return nil
}

// isDirectory takes a path strings and check the attributes if givin path
// is a dirctory or not
func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
