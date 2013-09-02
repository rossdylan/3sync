package main

import (
	"github.com/codegangsta/cli"
	"github.com/rossdylan/3sync/s3sync"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "3sync"
	app.Version = "0.1"
	app.Usage = "Go read the code"
	app.Commands = []cli.Command{
		{
			Name:      "sync",
			ShortName: "s",
			Action: func(c *cli.Context) {
				s3sync.Sync(c.GlobalString("path"), c.GlobalString("bucket"), c.GlobalString("region"), c.GlobalString("acl"))
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{"bucket", "", "What bucket to act on"},
		cli.StringFlag{"region", "us-east-1", "What AWS region to use"},
		cli.StringFlag{"acl", "private", "What S3 ACL to use"},
		cli.StringFlag{"path", "", "What local path to use"},
	}
	app.Run(os.Args)
}
