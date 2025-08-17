package main

import (
	"context"
	"flag"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/mrwormhole/laverna/synthesize"
)

var (
	filenamePath = flag.String("file", "", "filename path that is used for reading YAML file")
	maxWorkers   = flag.Int("workers", runtime.GOMAXPROCS(0), "maximum number of concurrent downloads")
)

func main() {
	flag.Parse()
	if *filenamePath == "" {
		flag.Usage()
		os.Exit(0)
	}

	raw, err := os.ReadFile(*filenamePath)
	if err != nil {
		log.Fatalf("failed to read file(%q): %v", *filenamePath, err)
	}

	isYAML := strings.HasSuffix(*filenamePath, ".yaml") || strings.HasSuffix(*filenamePath, ".yml")
	isCSV := strings.HasSuffix(*filenamePath, ".csv")
	if !isYAML && !isCSV {
		log.Fatalf("file format must be yaml/yml or csv")
	}

	var opts []synthesize.Opt
	if isYAML{
		opts, err = synthesize.UnmarshalYAML(raw)
		if err != nil {
			log.Fatalf("failed to unmarshal YAML: %v", err)
		}
	}
	if isCSV {
		opts, err = synthesize.UnmarshalCSV(raw)
		if err != nil {
			log.Fatalf("failed to unmarshal CSV: %v", err)
		}
	}

	runner := synthesize.NewBatchRunner(synthesize.WithMaxWorkers(*maxWorkers))
	if err := runner.Run(context.Background(), opts); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}
