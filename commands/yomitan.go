package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/occamist/laverna/yomitan"
)

type yomitanCmdFlags struct {
	Host  string
	Port  int
	HTTPS bool
}

func yomitanCmd(ctx context.Context, f yomitanCmdFlags) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := net.JoinHostPort(f.Host, strconv.Itoa(f.Port))
	server := &http.Server{
		Addr:              addr,
		Handler:           yomitan.NewHandler(http.DefaultClient, f.HTTPS),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %q", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("%T.ListenAndServe(): %w", server, err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("%T.Shutdown(): %w", server, err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

// NewYomitanCommand returns the "yomitan" command, which runs a local HTTP server
// implementing Yomitan's "Custom URL (JSON)" audio source.
func NewYomitanCommand() *cli.Command {
	return &cli.Command{
		Name:  "yomitan",
		Usage: "Runs a local HTTP server compatible with Yomitan's custom audio source",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "localhost",
				Usage: "`HOST` to bind the server to",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   8770,
				Usage:   "`PORT` to bind the server to",
			},
			&cli.BoolFlag{
				Name:  "https",
				Value: false,
				Usage: "generate audio URLs with an https:// scheme, e.g. when running behind a TLS-terminating reverse proxy",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return yomitanCmd(ctx, yomitanCmdFlags{
				Host:  cmd.String("host"),
				Port:  cmd.Int("port"),
				HTTPS: cmd.Bool("https"),
			})
		},
		Description: "Add http://localhost:8770?term={term}&reading={reading}&language={language} as a custom audio source (JSON) in Yomitan.\n\nlaverna yomitan --port 8770",
	}
}
