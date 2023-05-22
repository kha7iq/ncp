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
			Usage:   "UID is a globally applicable flag that can be utilized for write operations.",
		},
		&cli.IntFlag{
			Name:    "gid",
			Aliases: []string{"g"},
			Usage:   "GID is a globally applicable flag that can be utilized for write operations.",
		},
		&cli.BoolFlag{
			Name:    "turncate",
			Aliases: []string{"tr"},
			Usage:   "Enable or disable truncation of long file names in progress bar",
			Value:   true,
			EnvVars: []string{"NCP_FILENAME_TURNICATE"},
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
