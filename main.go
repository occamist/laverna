package main

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:                  "laverna",
		Description:           "Download Google Translate audios as mp3 files",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "path to config file",
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Value:   runtime.GOMAXPROCS(0),
				Usage:   "maximum number of concurrent downloads",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Downloads audios",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runCmd(ctx, cmd.String("file"), cmd.Int("workers"))
				},
			},
			{
				Name:  "anki",
				Usage: "Downloads audios to anki media folder and generates anki CSV file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "profile",
						Aliases: []string{"p"},
						Value:   "",
						Usage:   "anki profile name",
					},
					&cli.BoolFlag{ // to be passed down below to ankiCmd
						Name:    "shuffle",
						Aliases: []string{"s"},
						Value:   true,
						Usage:   "shuffles A,B,C,D choices per row",
					},
					&cli.BoolFlag{ // to be passed down below to ankiCmd
						Name:    "strip-csv-header",
						Aliases: []string{"strip"},
						Value:   true,
						Usage:   "strips csv header from the generated anki CSV file",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return ankiCmd(ctx, cmd.String("file"), cmd.Int("workers"), cmd.String("profile"))
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
