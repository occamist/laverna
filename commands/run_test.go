package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func writeTempFile(t *testing.T, content string, filename string) string {
	t.Helper()
	tmp := t.TempDir()
	fp := filepath.Join(tmp, filename)
	if err := os.WriteFile(fp, []byte(content), 0o600); err != nil { //nolint:gosec // always within t.TempDir() and OK for tests
		t.Fatalf("os.WriteFile(%q): %v", fp, err)
	}
	return fp
}

func TestRunCmd(t *testing.T) {
	t.Parallel()
	const maxWorkers = 20

	tests := []struct {
		name     string
		filename string
		setup    func(t *testing.T, filename string) string
		wantErr  error
	}{
		{
			name:     "successful run with CSV file",
			filename: "synthesize-example.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				raw, err := os.ReadFile("../testdata/synthesize-example.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
		},
		{
			name:     "invalid voice",
			filename: "synthesize-invalid-voice.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				raw, err := os.ReadFile("../testdata/synthesize-invalid-voice.csv")
				if err != nil {
					t.Fatalf("os.ReadFile(): %v", err)
				}
				return writeTempFile(t, string(raw), filename)
			},
			wantErr: errors.New("no audio line found"),
		},
		{
			name:     "empty CSV file",
			filename: "empty-synthesize.csv",
			setup: func(t *testing.T, filename string) string { //nolint:thelper // this is inline test helper
				return writeTempFile(t, "", filename)
			},
			wantErr: errors.New("empty csv"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.filename = tt.setup(t, tt.filename)

			f := runCmdFlags{
				Filename:   tt.filename,
				MaxWorkers: maxWorkers,
			}
			err := runCmd(t.Context(), f)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if !cmp.Equal(strings.Contains(err.Error(), tt.wantErr.Error()), true) {
					t.Errorf("runCmd(%v): wantErr=%v gotErr=%v", f, tt.wantErr, err)
				}
			}
		})
	}
}

func TestRunCmd_FileExtensions(t *testing.T) {
	const maxWorkers = 1

	tests := []struct {
		name     string
		filename string
		wantErr  error
	}{
		{
			name:     "invalid extension yaml",
			filename: "test.y4ml",
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
			f := runCmdFlags{
				Filename:   tt.filename,
				MaxWorkers: maxWorkers,
			}
			err := runCmd(t.Context(), f)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("runCmd(%v): wantErr=%v gotErr=%v", f, tt.wantErr, err)
				}
			}
		})
	}
}

func TestRunCmd_FileNotFound(t *testing.T) {
	const (
		maxWorkers = 1
		filename   = "nonexistent.csv"
	)

	f := runCmdFlags{
		Filename:   filename,
		MaxWorkers: maxWorkers,
	}
	err := runCmd(t.Context(), f)
	wantErr := errors.New(`failed to read file("nonexistent.csv"): open nonexistent.csv: no such file or directory`)
	if !cmp.Equal(wantErr, err, cmpopts.EquateErrors()) {
		if !cmp.Equal(wantErr.Error(), err.Error()) {
			t.Errorf("runCmd(%v): wantErr=%v gotErr=%v", f, wantErr, err)
		}
	}
}
