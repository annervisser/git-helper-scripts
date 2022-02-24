package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"pr-cli/commands"
)

func main() {
	app := &cli.App{
		Name:  "pr-cli",
		Usage: "make an explosive entrance",
		Commands: []*cli.Command{
			&commands.PickCmd,
		},
		EnableBashCompletion: true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
