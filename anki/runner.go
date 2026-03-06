package anki

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/mrwormhole/laverna/synthesize"
)

// MediaPath returns the Anki media path based on the profile name and the OS(runtime.GOOS)
func MediaPath(profile string, goos string) (string, error) {
	switch goos {
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
	profile    string
	save       SaveFn
}

// NewRunner creates a new Runner
func NewRunner(profile string, opts ...RunnerOption) (*Runner, error) {
	r := &Runner{
		client:     http.DefaultClient,
		maxWorkers: synthesize.DefaultMaxWorkers,
	}
	path, err := MediaPath(profile, runtime.GOOS)
	if err != nil {
		return nil, fmt.Errorf("ankiMediaPath(%q, %q): %w", profile, runtime.GOOS, err)
	}
	r.profile = profile

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

		// ensure directory exists
		if err := os.MkdirAll(path, 0o750); err != nil {
			return fmt.Errorf("failed to make directory(%q): %w", path, err)
		}

		_, err := os.Stat(fp)
		if err == nil { // err == nil means file exists
			return fmt.Errorf("saved file(%q) already exists", fp)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check file(%q): %w", fp, err)
		}

		return os.WriteFile(fp, audio, 0o600)
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

// RunConfig tells each run how to run
type RunConfig struct {
	Speed          string
	Voice          string
	HelperLanguage string
	Deck           string
	Endpoint       string
	Shuffle        bool
	StripCSVHeader bool
	PrintOut       bool
}

// Run runs given opts concurrently and stops if encounters an error
func (r *Runner) Run(ctx context.Context, reader io.Reader, c RunConfig) error {
	records, err := ReadCSVRecords(reader)
	if err != nil {
		return fmt.Errorf("ReadCSVRecords(): %w", err)
	}

	type job struct {
		opt      synthesize.Opt
		rowIndex int
		textType string
	}
	jobs := make(chan job, r.maxWorkers)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(jobs)
		for i, record := range records {
			type pair struct {
				text     string
				textType string
			}
			pairs := []pair{
				{record.CleanedText(), "AudioAnswer"},
				{record.TextA, "AudioA"},
				{record.TextB, "AudioB"},
				{record.TextC, "AudioC"},
				{record.TextD, "AudioD"},
			}

			for _, p := range pairs {
				opt := synthesize.Opt{
					Speed: synthesize.NewSpeed(c.Speed),
					Voice: synthesize.Voice(c.Voice),
					Text:  p.text,
				}
				select {
				case jobs <- job{opt: opt, rowIndex: i, textType: p.textType}:
				case <-gctx.Done():
					return gctx.Err()
				}
			}
		}
		return nil
	})

	const audioNum = 5 // There will be 5 audios per record
	type result struct {
		rowIndex int
		textType string
		uuid     string
	}
	results := make(chan result, len(records)*audioNum)

	for range r.maxWorkers {
		g.Go(func() error {
			for j := range jobs {
				if err := gctx.Err(); err != nil {
					return fmt.Errorf("%T.Err(): %w", gctx, err)
				}

				if strings.TrimSpace(j.opt.Text) == "" {
					return fmt.Errorf("text is empty on column(%q) and row(%d)", j.textType, j.rowIndex+1)
				}

				audio, err := synthesize.Run(gctx, r.client, j.opt)
				if err != nil {
					return fmt.Errorf("Run(%v) with row index(%d): %w", j.opt, j.rowIndex, err)
				}

				uuid := uuid.NewString()
				if err := r.save(uuid, audio); err != nil {
					return fmt.Errorf("%T.save(%v) with row index(%d): %w", r, uuid, j.rowIndex, err)
				}

				results <- result{rowIndex: j.rowIndex, textType: j.textType, uuid: uuid}
			}
			return nil
		})
	}

	collectedResults := make(map[int]map[string]string, len(records)) // rowIndex -> textType -> uuid
	done := make(chan struct{}, 1)

	go func() {
		defer close(done)
		for r := range results {
			if _, ok := collectedResults[r.rowIndex]; !ok {
				collectedResults[r.rowIndex] = make(map[string]string, audioNum)
			}
			collectedResults[r.rowIndex][r.textType] = r.uuid
		}
	}()

	if err := g.Wait(); err != nil {
		return fmt.Errorf("%T.Wait(): %w", g, err)
	}
	close(results)
	<-done

	for i := range records {
		uuids := collectedResults[i]

		records[i].AudioAnswer = "[sound:" + uuids["AudioAnswer"] + ".mp3]"
		records[i].AudioA = "[sound:" + uuids["AudioA"] + ".mp3]"
		records[i].AudioB = "[sound:" + uuids["AudioB"] + ".mp3]"
		records[i].AudioC = "[sound:" + uuids["AudioC"] + ".mp3]"
		records[i].AudioD = "[sound:" + uuids["AudioD"] + ".mp3]"
	}

	var writers []io.Writer
	var buf bytes.Buffer
	writers = append(writers, &buf)
	if c.PrintOut {
		writers = append(writers, os.Stdout)
	}
	multiWriter := io.MultiWriter(writers...)

	if err := WriteCSVRecords(multiWriter, records, c.StripCSVHeader, c.Shuffle); err != nil {
		return fmt.Errorf("WriteCSVRecords(): %w", err)
	}

	if err := r.postCSVRequest(ctx, c.Endpoint, c.Deck, bytes.NewReader(buf.Bytes())); err != nil {
		return fmt.Errorf("%T.postCSVRequest(%v, %v): %v", r, c.Endpoint, c.Deck, err)
	}
	return nil
}

func (r *Runner) postCSVRequest(ctx context.Context, endpoint, deck string, body io.Reader) error {
	URL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("url.Parse(%q): %v", endpoint, err)
	}
	q := URL.Query()
	q.Set("profile", r.profile)
	q.Set("deck", deck)
	URL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL.String(), body)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext(%v, %v): %v", http.MethodPost, URL.String(), err)
	}
	req.Header.Set("content-type", "text/csv")

	resp, err := r.client.Do(req) //nolint:gosec //endpoint is not a constant because it is configured by CLI flag
	if err != nil {
		return fmt.Errorf("%T.Do(): %w", r.client, err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("resp.Body.Close(): %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var em struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&em); err == nil {
			return fmt.Errorf("anki addon server returned message: %q", em.Message)
		}
	}
	return nil
}
