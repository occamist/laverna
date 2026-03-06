package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mrwormhole/laverna/anki"
	"github.com/mrwormhole/laverna/synthesize"
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

type ankiCmdFlags struct {
	Filename   string
	MaxWorkers int
	Profile    string
	Config     anki.RunConfig
}

func ankiCmd(ctx context.Context, f ankiCmdFlags) error {
	isCSV := strings.HasSuffix(f.Filename, ".csv")
	if !isCSV {
		return errors.New("file format must be csv")
	}

	file, err := os.Open(f.Filename)
	if err != nil {
		return fmt.Errorf("failed to open file(%q): %v", f.Filename, err)
	}
	defer func() { _ = file.Close() }()

	runner, err := anki.NewRunner(f.Profile, anki.WithMaxWorkers(f.MaxWorkers))
	if err != nil {
		return fmt.Errorf("failed to make runner: %v", err)
	}
	if err := runner.Run(ctx, file, f.Config); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
