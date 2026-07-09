package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/occamist/laverna/commands"
)

func main() {
	cmd := &cli.Command{
		Name:                  "laverna",
		Description:           "Download Google Translate audios as mp3 files",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			commands.NewRunCommand(),
			commands.NewAnkiCommand(),
			commands.NewYomitanCommand(),
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
