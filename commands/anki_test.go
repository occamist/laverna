package commands

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/occamist/laverna/anki"
)

func makeAnkiMediaPath(t *testing.T, profile string) {
	t.Helper()
	fp, err := anki.MediaPath(profile, runtime.GOOS)
	if err != nil {
		t.Fatalf("anki.MediaPath(%q, %q): %v", profile, runtime.GOOS, err)
	}
	if err := os.MkdirAll(fp, 0o750); err != nil {
		t.Fatalf("os.MkdirAll(%q): %v", fp, err)
	}
}

func TestAnkiCmd(t *testing.T) {
	t.Parallel()
	const maxWorkers = 20

	mockAnkiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() {
		mockAnkiServer.Close()
	})

	tests := []struct {
		name     string
		profile  string
		filename string
		setup    func(t *testing.T, filename string) string
		cfg      anki.RunConfig
		wantErr  error
	}{
		{
			name:     "successful run with default profile",
			profile:  "default",
			filename: "thai.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				makeAnkiMediaPath(t, "default")
				raw, err := os.ReadFile("../testdata/anki-th-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			cfg: anki.RunConfig{
				Speed:          "normal",
				Voice:          "th",
				Deck:           "test-deck",
				Endpoint:       mockAnkiServer.URL,
				Shuffle:        true,
				StripCSVHeader: true,
			},
		},
		{
			name:     "unknown voice in run config",
			profile:  "default",
			filename: "thai.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				makeAnkiMediaPath(t, "default")
				raw, err := os.ReadFile("../testdata/anki-th-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			cfg: anki.RunConfig{
				Speed:          "normal",
				Voice:          "XYZTESTXYZ",
				Shuffle:        true,
				StripCSVHeader: true,
			},
			wantErr: errors.New("no audio line found"),
		},
		{
			name:     "no anki media path",
			profile:  "unexistent profile",
			filename: "thai.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				raw, err := os.ReadFile("../testdata/anki-th-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			cfg: anki.RunConfig{
				Speed:          "normal",
				Voice:          "th",
				Shuffle:        true,
				StripCSVHeader: true,
			},
			wantErr: errors.New("anki media path"),
		},
		{
			name:     "empty CSV file",
			profile:  "default",
			filename: "thai.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				makeAnkiMediaPath(t, "default")
				return writeTempFile(t, "", filename)
			},
			cfg:     anki.RunConfig{},
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.filename = tt.setup(t, tt.filename)

			f := ankiCmdFlags{
				Filename:   tt.filename,
				MaxWorkers: maxWorkers,
				Profile:    tt.profile,
				Config:     tt.cfg,
			}
			err := ankiCmd(t.Context(), f)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if !cmp.Equal(strings.Contains(err.Error(), tt.wantErr.Error()), true) {
					t.Errorf("ankiCmd(%v): wantErr=%v gotErr=%v", f, tt.wantErr, err)
				}
			}
		})
	}
}

func TestAnkiCmd_FileExtensions(t *testing.T) {
	const (
		profile    = "default"
		maxWorkers = 1
	)

	tests := []struct {
		name     string
		filename string
		wantErr  error
	}{
		{
			name:     "invalid extension yaml",
			filename: "test.yaml",
			wantErr:  errors.New("file format must be csv"),
		},
		{
			name:     "no extension",
			filename: "test",
			wantErr:  errors.New("file format must be csv"),
		},
		{
			name:     "empty filename",
			filename: "",
			wantErr:  errors.New("file format must be csv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := ankiCmdFlags{
				Filename:   tt.filename,
				MaxWorkers: maxWorkers,
				Profile:    profile,
			}
			err := ankiCmd(t.Context(), f)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("ankiCmd(%v): wantErr=%v gotErr=%v", f, tt.wantErr, err)
				}
			}
		})
	}
}

func TestAnkiCmd_FileNotFound(t *testing.T) {
	const (
		profile    = "default"
		maxWorkers = 1
		filename   = "nonexistent.csv"
	)

	f := ankiCmdFlags{
		Filename:   filename,
		MaxWorkers: maxWorkers,
		Profile:    profile,
	}
	err := ankiCmd(t.Context(), f)
	wantErr := errors.New(`failed to open file("nonexistent.csv"): open nonexistent.csv: no such file or directory`)
	if !cmp.Equal(wantErr, err, cmpopts.EquateErrors()) {
		if !cmp.Equal(wantErr.Error(), err.Error()) {
			t.Errorf("ankiCmd(%v): wantErr=%v gotErr=%v", f, wantErr, err)
		}
	}
}
