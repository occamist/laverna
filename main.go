package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/mrwormhole/laverna/anki"
	"github.com/mrwormhole/laverna/synthesize"
)

func main() {
	cmd := &cli.Command{
		Name:                  "laverna",
		Description:           "Download Google Translate audios as mp3 files",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "filepath to prompt `FILE`",
				Required: true,
				Action: func(ctx context.Context, c *cli.Command, file string) error {
					if strings.TrimSpace(file) == "" {
						return errors.New("--file must not be blank")
					}
					return nil
				},
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
						Name:     "profile",
						Aliases:  []string{"p"},
						Usage:    "anki profile name",
						Required: true,
						Action: func(ctx context.Context, c *cli.Command, profile string) error {
							if strings.TrimSpace(profile) == "" {
								return errors.New("--profile must not be blank")
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:  "speed",
						Value: "normal",
						Usage: "specify the speed of audios, must be one of these values: `normal`, `slow`, `slowest`",
						Action: func(ctx context.Context, c *cli.Command, speed string) error {
							if !synthesize.IsSpeed(speed) {
								return errors.New("--speed must be one of these values: normal, slow, slowest")
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     "voice",
						Usage:    "specify the voice of audios",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "shuffle",
						Aliases: []string{"s"},
						Value:   true,
						Usage:   "shuffles A,B,C,D choices per row",
					},
					&cli.BoolFlag{
						Name:    "strip-csv-header",
						Aliases: []string{"strip"},
						Value:   true,
						Usage:   "strips csv header from the generated anki CSV file",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dir, filename := filepath.Dir(cmd.String("file")), "A"+filepath.Base(cmd.String("file"))
					outFilename := filepath.Join(dir, filename)

					cfg := anki.RunConfig{
						Speed:          cmd.String("speed"),
						Voice:          cmd.String("voice"),
						OutFilename:    outFilename,
						Shuffle:        cmd.Bool("shuffle"),
						StripCSVHeader: cmd.Bool("strip-csv-header"),
					}
					return ankiCmd(ctx, cmd.String("file"), cmd.Int("workers"), cmd.String("profile"), cfg)
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
