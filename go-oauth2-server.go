package main

import (
	"log"
	"os"

	"github.com/resonatecoop/id/cmd"
	"github.com/urfave/cli"
)

var (
	cliApp        *cli.App
	configBackend string
)

func init() {
	// Initialise a CLI app
	cliApp = cli.NewApp()
	cliApp.Name = "go-oauth2-server"
	cliApp.Usage = "Go OAuth 2.0 Server"
	cliApp.Author = "Richard Knop"
	cliApp.Email = "risoknop@gmail.com"
	cliApp.Version = "0.0.0"
	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "configBackend",
			Value:       "etcd",
			Destination: &configBackend,
		},
	}
}

func main() {
	// Set the CLI app commands
	cliApp.Commands = []cli.Command{
		{
			Name:  "runserver",
			Usage: "run web server",
			Action: func(c *cli.Context) error {
				return cmd.RunServer(configBackend)
			},
		},
	}

	// Run the CLI app
	if err := cliApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
