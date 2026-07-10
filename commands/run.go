package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/occamist/laverna/synthesize"
)

type runCmdFlags struct {
	Filename   string
	MaxWorkers int
}

func runCmd(ctx context.Context, f runCmdFlags) error {
	if !strings.HasSuffix(f.Filename, ".csv") {
		return errors.New("file format must be csv")
	}

	raw, err := os.ReadFile(f.Filename)
	if err != nil {
		return fmt.Errorf("failed to read file(%q): %v", f.Filename, err)
	}

	opts, err := synthesize.UnmarshalCSV(raw)
	if err != nil {
		return fmt.Errorf("failed to unmarshal CSV: %v", err)
	}

	runner := synthesize.NewRunner(synthesize.WithMaxWorkers(f.MaxWorkers))
	if err := runner.Run(ctx, opts); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}

// NewRunCommand returns the "run" command, which downloads audios described by a CSV file.
func NewRunCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Downloads audios",
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
				Value:   synthesize.DefaultMaxWorkers,
				Usage:   "maximum number of concurrent downloads",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runCmd(ctx, runCmdFlags{
				Filename:   cmd.String("file"),
				MaxWorkers: cmd.Int("workers"),
			})
		},
		Description: "laverna run --file example.csv",
	}
}
