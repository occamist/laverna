package commands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/net/proxy"

	"github.com/occamist/laverna/anki"
	"github.com/occamist/laverna/synthesize"
)

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
	if strings.TrimSpace(f.Proxy) != "" {
		proxyDialer, err := proxy.SOCKS5("tcp", f.Proxy, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS5 dialer(%q): %v", f.Proxy, err)
		}

		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, _ := net.SplitHostPort(addr)
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

// NewAnkiCommand returns the "anki" command, which downloads audios to the Anki media
// folder and generates an Anki-importable CSV file.
func NewAnkiCommand() *cli.Command {
	return &cli.Command{
		Name:  "anki",
		Usage: "Downloads audios to anki media folder and generates anki CSV file",
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
				Name:     "deck",
				Aliases:  []string{"d"},
				Value:    "laverna-deck",
				Usage:    "anki deck name",
				Required: true,
				Action: func(ctx context.Context, c *cli.Command, deck string) error {
					if strings.TrimSpace(deck) == "" {
						return errors.New("--deck must not be blank")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "endpoint",
				Aliases: []string{"e"},
				Value:   "http://localhost:5555/v1/import-csv",
				Usage:   "anki addon endpoint `URL`",
			},
			&cli.StringFlag{
				Name:    "speed",
				Aliases: []string{"s"},
				Value:   "normal",
				Usage:   "specify the `SPEED` of audios",
				Action: func(ctx context.Context, c *cli.Command, speed string) error {
					if !synthesize.IsSpeed(speed) {
						return errors.New("--speed must be one of these values: normal, slow, slowest")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:     "voice",
				Aliases:  []string{"v"},
				Usage:    "specify the `VOICE` of audios",
				Required: true,
				Action: func(ctx context.Context, c *cli.Command, voice string) error {
					if !synthesize.IsVoice(voice) {
						return errors.New("--voice must be one of these values: " + strings.Join(synthesize.AllVoices, ", "))
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "shuffle",
				Value: true,
				Usage: "shuffles the text choices per row",
			},
			&cli.BoolFlag{
				Name:  "strip-csv-header",
				Value: true,
				Usage: "strips the csv header from the generated anki CSV file",
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Value: false,
				Usage: "prints the generated anki CSV file to stdout",
			},
			&cli.StringFlag{
				Name:  "proxy",
				Usage: "SOCKS5 proxy `ADDRESS` (e.g. localhost:1080)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return ankiCmd(ctx, ankiCmdFlags{
				Filename:   cmd.String("file"),
				MaxWorkers: cmd.Int("workers"),
				Profile:    cmd.String("profile"),
				Proxy:      cmd.String("proxy"),
				Config: anki.RunConfig{
					Speed:          cmd.String("speed"),
					Voice:          cmd.String("voice"),
					Deck:           cmd.String("deck"),
					Endpoint:       cmd.String("endpoint"),
					Shuffle:        cmd.Bool("shuffle"),
					StripCSVHeader: cmd.Bool("strip-csv-header"),
					PrintOut:       cmd.Bool("stdout"),
				},
			})
		},
		Description: "Anki app has to be launched locally with Laverna Anki plugin installed before running CLI commands.\n\nlaverna anki --profile my-profile --deck my-deck --voice en --file example.csv",
	}
}
