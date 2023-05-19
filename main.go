package main

import (
	"log"
	"os"

	"github.com/kha7iq/ncp/cmd/from"
	"github.com/kha7iq/ncp/cmd/to"
	"github.com/urfave/cli/v2"
)

// Version variables are used for semVer
var (
	version   string
	buildDate string
	commitSHA string
)

// main with all the function into commands
func main() {
	app := cli.NewApp()
	app.Name = "ncp"
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "uid",
			Aliases: []string{"u"},
			Usage:   "UID can be utilized to perform the write operations.",
		},
		&cli.IntFlag{
			Name:    "gid",
			Aliases: []string{"g"},
			Usage:   "GID can be utilized to perform the write operations.",
		},
	}
	app.Version = version + " BuildDate: " + buildDate + " " + " CommitSHA: " + commitSHA
	app.Usage = "provides a straightforward and efficient way to handle file transfers between the local machine and a NFS server."
	app.Description = `NCP is used to efficiently copy files to and from an NFS server.
It provides a convenient way to transfer files between the local machine and the NFS server,
supporting both upload and download operations.`
	app.Commands = []*cli.Command{
		to.ToServer(),
		from.FromServer(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
