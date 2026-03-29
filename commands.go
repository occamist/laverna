package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/proxy"

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
	Proxy      string
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

	opts := []anki.RunnerOption{anki.WithMaxWorkers(f.MaxWorkers)}
	if f.Proxy != "" {
		proxyDialer, err := proxy.SOCKS5("tcp", f.Proxy, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS5 dialer(%q): %v", f.Proxy, err)
		}

		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, _ := net.SplitHostPort(addr)
				fmt.Printf("host: %v, network: %v\n", host, network)
				if host == "localhost" || host == "127.0.0.1" || host == "::1" {
					var d net.Dialer
					return d.DialContext(ctx, network, addr)
				}
				return proxyDialer.Dial(network, addr)
			},
		}
		opts = append(opts, anki.WithClient(&http.Client{Transport: transport}))
	}

	runner, err := anki.NewRunner(f.Profile, opts...)
	if err != nil {
		return fmt.Errorf("failed to make runner: %v", err)
	}
	if err := runner.Run(ctx, file, f.Config); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
