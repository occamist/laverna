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
		log.Fatalf("[ERR] failed to read filename path: %v", err)
	}

	var opts []synthesize.Opt
	if strings.HasSuffix(*filenamePath, ".yaml") || strings.HasSuffix(*filenamePath, ".yml") {
		opts, err = synthesize.UnmarshalYAML(raw)
		if err != nil {
			log.Fatalf("[ERR] failed to unmarshal YAML: %v", err)
		}
	} else if strings.HasSuffix(*filenamePath, ".csv") {
		opts, err = synthesize.UnmarshalCSV(raw)
		if err != nil {
			log.Fatalf("[ERR] failed to unmarshal CSV: %v", err)
		}
	} else {
		log.Fatalf("[ERR] file format must be yaml/yml or csv")
		return
	}

	runner := synthesize.NewBatchRunner(synthesize.WithMaxWorkers(*maxWorkers))
	if err := runner.Run(context.Background(), opts); err != nil {
		log.Fatalf("[ERR] failed to run batch: %v", err)
	}
}
