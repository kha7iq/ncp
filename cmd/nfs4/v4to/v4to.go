package v4to

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
	inputPath      string
	nfsHost        string
	nfsMountFolder string
	nfsServerPort  string
}

type progressReader struct {
	reader io.Reader
	bar    *progressbar.ProgressBar
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.bar.Add(n)
	return n, err
}

// ToServer function provides functionaltiy to transfer files or folders from local filesystem to NFS server.
func ToServerV4() *cli.Command {
	var nc nfsConfg
	return &cli.Command{
		Name:      "v4to",
		Usage:     "The 'v4to' command is used to copy files or folders from the local machine to NFS v4 server.",
		UsageText: "ncp v4to --host 192.168.0.80 --nfspath data --input src",
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
				Usage:       "IP address or hostname that can be used to access the NFS server.",
			},
			&cli.StringFlag{
				Destination: &nc.nfsMountFolder,
				Required:    true,
				Name:        "nfspath",
				Aliases:     []string{"p"},
				Usage:       "NFS path denotes the destination directory on the NFS server where files will be copied to.",
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
			// context for nfs4 client
			ctxx := context.Background()
			hostNameLocal, _ := os.Hostname()

			nfs4, err := nfs4.NewNfsClient(ctxx, nc.nfsHost+":"+nc.nfsServerPort, nfs4.AuthParams{
				MachineName: hostNameLocal,
				Uid:         uid,
				Gid:         gid,
			})
			if err != nil {
				return err
			}
			defer nfs4.Close()

			_, err = helper.IsPathValid(nc.inputPath)
			if err != nil {
				log.Fatalf("input path error: %v", err)
			}

			basePath := filepath.Dir(nc.inputPath)
			folders, files, err := getFolderAndFileList(nc.inputPath, "")
			if err != nil {
				log.Fatalf("unable to get list of files and folders %v", err)
			}
			if isDirectory(nc.inputPath) {

				for _, v := range folders {
					targetDir := nc.nfsMountFolder + "/" + v
					err = nfs4.MakePath(targetDir)
					if err != nil {
						return err // But return all other errors
					}
				}
				for _, sourcFile := range files {
					targetfile := nc.nfsMountFolder + "/" + sourcFile
					sf := filepath.Join(basePath, sourcFile)
					// Copy file to destination
					if err = transferFile(nfs4, sf, targetfile, truncate); err != nil {
						log.Fatalf("fail to transfer files %v", err)
					}
				}
			} else {
				nfs4.MakePath(nc.nfsMountFolder)
				for _, sourcFile := range files {
					targetfile := nc.nfsMountFolder + "/" + sourcFile
					sf := filepath.Join(basePath, sourcFile)
					// Copy file to destination
					if err = transferFile(nfs4, sf, targetfile, truncate); err != nil {
						log.Fatalf("fail to transfer files %v", err)
					}
				}
			}
			return nil
		},
	}
}

// listFileAndFolders take a directory path and returns a slice containng files and another containing folders
func getFolderAndFileList(path string, basePath string) ([]string, []string, error) {
	var folders []string
	var files []string

	// Check if the path is a directory
	isDir := isDirectory(path)

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
		isDir := isDirectory(contentPath)
		if isDir {
			subfolderFolders, subfolderFiles, err := getFolderAndFileList(contentPath, filepath.Join(basePath, folderName))
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

// transferFile will take a source and target file path along with *nfs4.NfsClient to transfer file
func transferFile(nfs4 *nfs4.NfsClient, srcfile string, targetfile string, turnication bool) error {
	var filePath string
	sourceFile, err := os.Open(srcfile)
	if err != nil {
		log.Fatalf("error opening source file: %s", err.Error())
	}
	defer sourceFile.Close()

	// Get file information to obtain its size
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	fileSize := fileInfo.Size()

	if !turnication {
		filePath = srcfile
	} else {
		filePath = helper.TruncateFileName(srcfile)
	}
	// Create a progress bar based on the file size
	bar := helper.ProgressBar(fileSize, filePath, helper.CheckMark())
	// bar := progressbar.DefaultBytes(fileSize, "Copying")

	// Create a progress reader that wraps the source file reader
	reader := &progressReader{
		reader: sourceFile,
		bar:    bar,
	}

	// Call the nfs4.WriteFile function with the correct arguments
	_, err = nfs4.WriteFile(targetfile, true, 0, reader)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	bar.Finish()

	return nil
}

// isDirectory takes a path strings and check the attributes if givin path
// is a dirctory or not
func isDirectory(path string) bool {
	info, _ := os.Stat(path)
	return info.IsDir()
}
