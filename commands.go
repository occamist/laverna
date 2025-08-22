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

func runCmd(ctx context.Context, filename string, maxWorkers int) error {
	if filename == "" {
		return errors.New("--file must not be blank")
	}

	isYAML := strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
	isCSV := strings.HasSuffix(filename, ".csv")
	if !isYAML && !isCSV {
		return errors.New("file format must be yaml/yml or csv")
	}

	raw, err := os.ReadFile(filename) //nolint:gosec // file inclusion is intended
	if err != nil {
		return fmt.Errorf("failed to read file(%q): %v", filename, err)
	}

	var opts []synthesize.Opt
	if isYAML {
		opts, err = synthesize.UnmarshalYAML(raw)
		if err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %v", err)
		}
	}
	if isCSV {
		opts, err = synthesize.UnmarshalCSV(raw)
		if err != nil {
			return fmt.Errorf("failed to unmarshal CSV: %v", err)
		}
	}

	runner := synthesize.NewRunner(synthesize.WithMaxWorkers(maxWorkers))
	if err := runner.Run(ctx, opts); err != nil {
		return fmt.Errorf("failed to run: %v", err)
	}

	return nil
}

func ankiCmd(ctx context.Context, filename string, maxWorkers int, profile string) error {
	if filename == "" {
		return errors.New("--file must not be blank")
	}
	if profile == "" {
		return errors.New("--profile must not be blank")
	}

	isCSV := strings.HasSuffix(filename, ".csv")
	if !isCSV {
		return errors.New("file format must be csv")
	}

	f, err := os.Open(filename) //nolint:gosec // file inclusion is intended
	if err != nil {
		return fmt.Errorf("failed to open file(%q): %v", filename, err)
	}
	defer func() {
		_ = f.Close()
	}()

	runner, err := anki.NewRunner(profile, anki.WithMaxWorkers(maxWorkers))
	if err != nil {
		return fmt.Errorf("failed to make runner: %v", err)
	}

	outFilename := "A" + filename
	if err := runner.Run(ctx, f, outFilename); err != nil {
		return fmt.Errorf("failed to run: %v", err)
	}

	return nil
}
