package from

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

type nfsConfg struct {
	nfsHost        string
	nfsMountFolder string
}

// FromServer function provides functionaltiy to transfer files or folders from NFS server to local filesystem.
func FromServer() *cli.Command {
	var nc nfsConfg
	return &cli.Command{
		Name:      "from",
		Usage:     "The 'from' command is used to copy files or folders from Remote NFS server to local machine.",
		UsageText: "ncp to --host 192.168.0.80 --nfspath data/tee",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &nc.nfsHost,
				Name:        "host",
				Aliases:     []string{"t"},
				Required:    true,
				Usage:       "The IP address or DNS of the NFS server specifies the IP or hostname that can be used to access the NFS server.",
			},
			&cli.StringFlag{
				Destination: &nc.nfsMountFolder,
				Name:        "nfspath",
				Required:    true,
				Aliases:     []string{"p"},
				Usage:       "The NFS path denotes the destination directory on the NFS server where files/folders will be copied from.",
			},
		},
		Action: func(ctx *cli.Context) error {
			u := ctx.Int("uid")
			g := ctx.Int("gid")
			uid, gid := checkUID(u, g)

			rootDir := filepath.Dir(nc.nfsMountFolder)
			dir := filepath.Base(nc.nfsMountFolder)

			mount, err := nfs.DialMount(nc.nfsHost, false)
			if err != nil {
				log.Fatalf("unable to dial MOUNT service: %v", err)
			}
			defer mount.Close()

			hostNameLocal, _ := os.Hostname()

			auth := rpc.NewAuthUnix(hostNameLocal, uid, gid)

			nfs, err := mount.Mount(rootDir, auth.Auth())
			if err != nil {
				log.Fatalf("unable to mount volume: %v", err)
			}
			defer nfs.Close()

			if err = mount.Unmount(); err != nil {
				log.Fatalf("unable to unmount target: %v", err)
			}

			mount.Close()

			if isDirectory(nfs, dir) {

				dirs, files, err := listFilesAndFolders(nfs, dir)
				if err != nil {
					log.Fatalf("failed to list files and folders: %v", err)
				}
				for _, v := range dirs {
					if err = createDirIfNotExist(v); err != nil {
						return err
					}
				}

				for _, sf := range files {
					if err = transferFile(nfs, sf, sf); err != nil {
						log.Fatalf("fail")

					}
				}

			}

			if !isDirectory(nfs, dir) {

				if err = transferFile(nfs, dir, dir); err != nil {
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

// transferFile will take a source file path and target file path and transfer file
func transferFile(nfs *nfs.Target, srcfile string, targetfile string) error {

	sourceFile, err := nfs.Open(srcfile)
	if err != nil {
		util.Errorf("error opening source file: %s", err.Error())
		return err
	}

	// Calculate the ShaSum
	h := sha256.New()
	t := io.TeeReader(sourceFile, h)
	stat, _, _ := nfs.Lookup(srcfile)
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
		progressbar.OptionSetWidth(20),
		progressbar.OptionShowBytes(true),
		progressbar.OptionOnCompletion(checkMark()),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSpinnerType(14),
	)

	wr, err := os.Create(targetfile)
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

// createDirIfNotExist will check if folder does not exist and automatically creates it.
func createDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
		return err
	}
	return nil
}

// listFileAndFolders take a directory path and returns a slice containng files and another containing folders
func listFilesAndFolders(v *nfs.Target, dir string) ([]string, []string, error) {
	outDirs, err := v.ReadDirPlus(dir)
	if err != nil {
		return nil, nil, err
	}

	var dirs []string
	var files []string

	for _, outDir := range outDirs {
		if outDir.Name() != "." && outDir.Name() != ".." {
			if outDir.IsDir() {
				subDirs, subFiles, err := listFilesAndFolders(v, dir+"/"+outDir.Name())
				if err != nil {
					return nil, nil, err
				}
				dirs = append(dirs, subDirs...)
				files = append(files, subFiles...)
				dirs = append(dirs, dir+"/"+outDir.Name())
			} else {
				files = append(files, dir+"/"+outDir.Name())
			}
		}
	}

	return dirs, files, nil
}

// isDirectory takes a path strings and check the attirbutes if givin path
// is a dirctory or not
func isDirectory(v *nfs.Target, dir string) bool {
	outDirs, _ := v.ReadDirPlus(dir)
	for _, outDir := range outDirs {
		if outDir.Name() != "." && outDir.Name() != ".." {
			if outDir.IsDir() {
				return true
			}
		}
	}

	return false
}

func checkUID(u int, g int) (uid, gid uint32) {
	if u == 0 {
		uid = uint32(0)
		gid = uint32(0)
	} else {
		uid = uint32(u)
		gid = uint32(g)
	}
	return uid, gid
}
