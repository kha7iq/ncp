package v4from

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/kha7iq/go-nfs-client/nfs4"
	"github.com/kha7iq/ncp/internal/helper"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

type nfsConfg struct {
	nfsHost        string
	nfsMountFolder string
	nfsServerPort  string
}

type progressWriter struct {
	writer io.Writer
	bar    *progressbar.ProgressBar
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.bar.Add(n)
	return n, err
}

// FromServer function provides functionaltiy to transfer files or folders from NFS server to local filesystem.
func FromServerV4() *cli.Command {
	var nc nfsConfg
	return &cli.Command{
		Name:      "v4from",
		Usage:     "The 'v4from' command is used to copy files or folders from Remote NFS v4 server to local machine.",
		UsageText: "ncp v4from --host 192.168.0.80 --nfspath data/src",
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
			&cli.StringFlag{
				Destination: &nc.nfsServerPort,
				Name:        "port",
				Aliases:     []string{"pr"},
				Usage:       "NFS server port, if other then default.",
				Value:       "2049",
			},
		},
		Action: func(ctx *cli.Context) error {
			truncate := ctx.Bool("turncate")
			u := ctx.Int("uid")
			g := ctx.Int("gid")
			uid, gid := helper.CheckUID(u, g)
			ctxx := context.Background()
			hostNameLocal, _ := os.Hostname()
			basePath := filepath.Base(nc.nfsMountFolder)

			nfs4, err := nfs4.NewNfsClient(ctxx, nc.nfsHost+":"+nc.nfsServerPort, nfs4.AuthParams{
				MachineName: hostNameLocal,
				Uid:         uid,
				Gid:         gid,
			})
			if err != nil {
				return err
			}
			defer nfs4.Close()

			if isDirectory(nfs4, nc.nfsMountFolder) {

				folders, files, err := getFolderAndFileList(nfs4, nc.nfsMountFolder)
				if err != nil {
					log.Fatalf("unable to get list of files and folders %v", err)
				}
				if len(folders) == 0 {
					folders = append(folders, nc.nfsMountFolder)
				}

				for _, v := range folders {
					if err = createDirIfNotExist(v); err != nil {
						log.Fatalf("fail to create folder %V", err)
					}
				}

				for _, sf := range files {

					if err = transferFile(nfs4, sf, sf, truncate); err != nil {
						log.Fatalf("fail to copy files with error %V", err)
					}
				}
			}
			if !isDirectory(nfs4, nc.nfsMountFolder) {
				if err = transferFile(nfs4, nc.nfsMountFolder, basePath, truncate); err != nil {
					log.Fatalf("fail to transfer files %v", err)
				}
			}
			return nil
		},
	}
}

// transferFile will take a source and target file path along with nfs4.NfsInterface to transfer file
func transferFile(nfs4 nfs4.NfsInterface, srcfile string, targetfile string, truncate bool) error {
	var filePath string

	st, err := nfs4.GetFileInfo(srcfile)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}
	fileSize := st.Size

	if !truncate {
		filePath = srcfile
	} else {
		filePath = helper.TruncateFileName(srcfile)
	}

	wr, err := os.Create(targetfile)
	if err != nil {
		log.Fatalf("error opening target file: %s", err.Error())
		return err
	}
	defer wr.Close()

	progress := helper.ProgressBar(int64(fileSize), filePath, helper.CheckMark())

	// Create a writer with a progress callback to update the progress bar
	writer := &progressWriter{
		writer: wr,
		bar:    progress,
	}
	// Copy files with progress size
	_, err = nfs4.ReadFile(srcfile, 0, uint64(fileSize), writer)
	if err != nil {
		return fmt.Errorf("failed to read remote file: %w", err)
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
func getFolderAndFileList(nfs4 nfs4.NfsInterface, remotePath string) ([]string, []string, error) {

	entries, err := nfs4.GetFileList(remotePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve remote file list: %w", err)
	}

	var folders []string
	var files []string

	for _, entry := range entries {
		if entry.IsDir {
			// Add folder to the list
			subfolders, subfiles, err := getFolderAndFileList(nfs4, remotePath+"/"+entry.Name)
			if err != nil {
				return nil, nil, err
			}

			folders = append(folders, subfolders...)
			files = append(files, subfiles...)
			folders = append(folders, remotePath+"/"+entry.Name)
		} else {
			// Add file to the list
			files = append(files, remotePath+"/"+entry.Name)
		}
	}

	return folders, files, nil
}

// isDir takes a path strings and check the attributes if givin path
// is a dirctory or not
func isDirectory(nfs4 nfs4.NfsInterface, remotePath string) bool {
	fileInfo, _ := nfs4.GetFileInfo(remotePath)
	return fileInfo.IsDir
}
