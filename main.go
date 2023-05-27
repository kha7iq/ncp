package main

import (
	"log"
	"os"

	"github.com/kha7iq/ncp/cmd/nfs3/from"
	"github.com/kha7iq/ncp/cmd/nfs3/to"
	"github.com/kha7iq/ncp/cmd/nfs4/v4from"
	"github.com/kha7iq/ncp/cmd/nfs4/v4to"
	"github.com/kha7iq/ncp/internal/helper"
	"github.com/urfave/cli/v2"
)

// Version variables are used for semVer
var (
	version   string
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
	app.Version = version + " CommitSHA: " + helper.TrimSHA(commitSHA)
	app.Usage = "provides a straightforward and efficient way to handle file transfers between the local machine and a NFS server."
	app.Description = `NCP offers a user-friendly solution for efficiently transferring files and folders between your local machine
and the NFS server. It enables seamless recursive upload and download operations, supporting both NFS v3 and NFS V4 protocols.`
	app.Commands = []*cli.Command{
		to.ToServerV3(),
		from.FromServerV3(),
		v4to.ToServerV4(),
		v4from.FromServerV4(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
