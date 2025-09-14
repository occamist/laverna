package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/mrwormhole/laverna/anki"
)

func writeTempFile(t *testing.T, content string, filename string) string {
	t.Helper()
	tmp := t.TempDir()
	fp := filepath.Join(tmp, filename)
	if err := os.WriteFile(fp, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", fp, err)
	}
	return fp
}

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
	const (
		maxWorkers = 20
	)

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
				raw, err := os.ReadFile("testdata/anki-th-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			cfg: anki.RunConfig{
				Speed:          "normal",
				Voice:          "th",
				OutFilename:    filepath.Join(t.TempDir(), "Athai.csv"),
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
				raw, err := os.ReadFile("testdata/anki-th-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			cfg: anki.RunConfig{
				Speed:          "normal",
				Voice:          "XYZTESTXYZ",
				OutFilename:    filepath.Join(t.TempDir(), "unknown.csv"),
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
				raw, err := os.ReadFile("testdata/anki-th-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			cfg: anki.RunConfig{
				Speed:          "normal",
				Voice:          "th",
				OutFilename:    filepath.Join(t.TempDir(), "Athai.csv"),
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
			err := ankiCmd(t.Context(), tt.filename, maxWorkers, tt.profile, tt.cfg)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if !cmp.Equal(strings.Contains(err.Error(), tt.wantErr.Error()), true) {
					t.Errorf("ankiCmd(%q, %d, %q, %v): wantErr=%v gotErr=%v", tt.filename, maxWorkers, tt.profile, tt.cfg, tt.wantErr, err)
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
			err := ankiCmd(t.Context(), tt.filename, maxWorkers, profile, anki.RunConfig{})

			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("ankiCmd(%q, %d, %q, cfg): wantErr=%v gotErr=%v", tt.filename, maxWorkers, profile, tt.wantErr, err)
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
	err := ankiCmd(t.Context(), filename, maxWorkers, profile, anki.RunConfig{})

	wantErr := errors.New(`failed to open file("nonexistent.csv"): open nonexistent.csv: no such file or directory`)
	if !cmp.Equal(wantErr, err, cmpopts.EquateErrors()) {
		if !cmp.Equal(wantErr.Error(), err.Error()) {
			t.Errorf("ankiCmd(%q, %d, %q): wantErr=%v gotErr=%v", filename, maxWorkers, profile, wantErr, err)
		}
	}
}
