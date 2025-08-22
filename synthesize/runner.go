package synthesize

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/sourcegraph/conc/pool"
)

// SaveFn is function that allows additional save step for BatchRunner's per run
type SaveFn func(filename string, audio []byte) error

// Runner handles concurrent processing of synthesize operations
type Runner struct {
	client     *http.Client
	maxWorkers int
	save       SaveFn
}

// NewRunner creates a new Runner with the given options
func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		client:     http.DefaultClient,
		maxWorkers: runtime.GOMAXPROCS(0),
		save: func(filename string, audio []byte) error {
			return os.WriteFile(filename+".mp3", audio, 0600)
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
	p := pool.New().WithContext(ctx).WithMaxGoroutines(r.maxWorkers)

	for _, opt := range opts {
		p.Go(func(ctx context.Context) error {
			audio, err := Run(ctx, r.client, opt)
			if err != nil {
				return fmt.Errorf("Run(%v): %w", opt, err)
			}

			if err := r.save(opt.Text, audio); err != nil {
				return fmt.Errorf("%T.SaveFunc(%v): %w", p, opt.Text, err)
			}
			return nil
		})
	}

	if err := p.Wait(); err != nil {
		return fmt.Errorf("%T.Wait(): %w", p, err)
	}
	return nil
}
