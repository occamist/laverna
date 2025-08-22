package anki

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"

	"github.com/mrwormhole/laverna/synthesize"
)

func ankiMediaPath(profile string) (string, error) {
	switch runtime.GOOS {
	case "windows": // %APPDATA%\Anki2\<profile>\collection.media
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New("APPDATA env variable is missing")
		}
		return filepath.Join(appData, "Anki2", profile, "collection.media"), nil
	case "darwin": // ~/Library/Application Support/Anki2/<profile>/collection.media
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		return filepath.Join(dir, "Library", "Application Support", "Anki2", profile, "collection.media"), nil
	case "linux": // ~/.local/share/Anki2/<profile>/collection.media
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		return filepath.Join(dir, ".local", "share", "Anki2", profile, "collection.media"), nil
	default:
		return "anki", nil
	}
}

// SaveFn is function that allows additional save step for Runner's per run
type SaveFn func(filename string, audio []byte) error

// Runner handles concurrent processing of synthesize operations
type Runner struct {
	client     *http.Client
	maxWorkers int
	save       SaveFn
}

// NewRunner creates a new Runner
func NewRunner(profile string, opts ...RunnerOption) (*Runner, error) {
	r := &Runner{
		client:     http.DefaultClient,
		maxWorkers: runtime.GOMAXPROCS(0),
	}
	path, err := ankiMediaPath(profile)
	if err != nil {
		return nil, fmt.Errorf("ankiMediaPath(%q): %w", profile, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("anki media path(%q) does not exist", path)
		}
		return nil, fmt.Errorf("os.Stat(path): %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("anki media path(%q) must be a directory", path)
	}

	saveFile := func(filename string, audio []byte) error {
		filename = filename + ".mp3"
		fp := filepath.Join(path, filename)

		info, err := os.Stat(fp)
		if info.Size() > 0 && err == nil {
			return fmt.Errorf("saved file(%q) already exists", fp)
		}

		return os.WriteFile(fp, audio, 0600)
	}
	r.save = saveFile

	for _, opt := range opts {
		opt(r)
	}
	return r, nil
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
func (r *Runner) Run(ctx context.Context, reader io.Reader, outFilename string) error {
	p := pool.New().WithContext(ctx).WithMaxGoroutines(r.maxWorkers)

	records, err := ReadCSVRecords(reader)
	if err != nil {
		return fmt.Errorf("readRecords(): %w", err)
	}

	// speed must be passed as a flag
	// voice is only different for HelperText
	// There will be 6 texts per record
	for _, record := range records {
		p.Go(func(ctx context.Context) error {
			opt := synthesize.Opt{
				Speed: synthesize.NormalSpeed,
				Voice: synthesize.NepaliVoice,
				Text:  record.Text,
			}
			audio, err := synthesize.Run(ctx, r.client, opt)
			if err != nil {
				return fmt.Errorf("Run(%v): %w", opt, err)
			}

			uuid := uuid.NewString()
			if err := r.save(uuid, audio); err != nil {
				return fmt.Errorf("%T.SaveFunc(%v): %w", p, uuid, err)
			}
			return nil
		})
	}

	if err := p.Wait(); err != nil {
		return fmt.Errorf("%T.Wait(): %w", p, err)
	}
	return nil
}
