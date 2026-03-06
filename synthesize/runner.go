package synthesize

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

// SaveFn is function that allows additional save step for BatchRunner's per run
type SaveFn func(filename string, audio []byte) error

// Runner handles concurrent processing of synthesize operations
type Runner struct {
	client     *http.Client
	maxWorkers int
	save       SaveFn
}

// DefaultMaxWorkers is the number of concurrent requests for scraping the audios
const DefaultMaxWorkers = 100

// NewRunner creates a new Runner with the given options
func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		client:     http.DefaultClient,
		maxWorkers: DefaultMaxWorkers,
		save: func(filename string, audio []byte) error {
			return os.WriteFile(filename+".mp3", audio, 0o600)
		},
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RunnerOption configures a Runner
type RunnerOption func(*Runner)

// WithClient sets the HTTP client
func WithClient(c *http.Client) RunnerOption {
	return func(r *Runner) {
		r.client = c
	}
}

// WithMaxWorkers sets the maximum number of concurrent workers
func WithMaxWorkers(n int) RunnerOption {
	return func(r *Runner) {
		r.maxWorkers = n
	}
}

// WithSaveFunc sets custom save function
func WithSaveFunc(fn SaveFn) RunnerOption {
	return func(r *Runner) {
		r.save = fn
	}
}

// Run runs given opts concurrently and stops if encounters an error
func (r *Runner) Run(ctx context.Context, opts []Opt) error {
	jobs := make(chan Opt, r.maxWorkers)
	g, ctx := errgroup.WithContext(ctx)

	go func() {
		defer close(jobs)
		for _, opt := range opts {
			select {
			case jobs <- opt:
			case <-ctx.Done():
				return
			}
		}
	}()

	for range r.maxWorkers {
		g.Go(func() error {
			for j := range jobs {
				audio, err := Run(ctx, r.client, j)
				if err != nil {
					return fmt.Errorf("Run(%v): %w", j, err)
				}
				if err := r.save(j.Text, audio); err != nil {
					return fmt.Errorf("%T.save(%v): %w", r, j.Text, err)
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("%T.Wait(): %w", g, err)
	}
	return nil
}
