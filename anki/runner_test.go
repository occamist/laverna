package anki

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAnkiMediaPath(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		profile  string
		setupEnv func(t *testing.T)
		want     string
		wantErr  error
	}{
		{
			name:    "windows missing APPDATA",
			goos:    "windows",
			profile: "User 1",
			wantErr: errors.New("APPDATA env variable is missing"),
		},
		{
			name:    "windows empty APPDATA",
			goos:    "windows",
			profile: "User 1",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("APPDATA", "")
			},
			wantErr: errors.New("APPDATA env variable is missing"),
		},
		{
			name:    "darwin",
			goos:    "darwin",
			profile: "User 1",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("HOME", "/Users/test")
			},
			want: "/Users/test/Library/Application Support/Anki2/User 1/collection.media",
		},
		{
			name:    "linux",
			goos:    "linux",
			profile: "User 1",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("HOME", "/home/test")
			},
			want: "/home/test/.local/share/Anki2/User 1/collection.media",
		},
		{
			name:    "unknown os",
			goos:    "freebsd",
			profile: "User 1",
			want:    "anki",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv(t)
			}

			got, err := MediaPath(tt.profile, tt.goos)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if !cmp.Equal(err.Error(), tt.wantErr.Error()) {
					t.Errorf("ankiMediaPath(%q, %q): wantErr=%v gotErr=%v", tt.profile, tt.goos, tt.wantErr, err)
				}
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ankiMediaPath(%q, %q) mismatch (-want +got):\n%s", tt.profile, tt.goos, diff)
			}
		})
	}
}

func TestNewRunner(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		setup   func(t *testing.T)
		wantErr string
	}{
		{
			name:    "success",
			profile: "default",
			setup: func(t *testing.T) {
				t.Helper()
				home := t.TempDir()
				t.Setenv("HOME", home)
				mediaPath := filepath.Join(home, ".local", "share", "Anki2", "default", "collection.media")
				if err := os.MkdirAll(mediaPath, 0o750); err != nil {
					t.Fatalf("os.MkdirAll(%q): %v", mediaPath, err)
				}
			},
		},
		{
			name:    "media path does not exist",
			profile: "missing-profile",
			setup: func(t *testing.T) {
				t.Helper()
				t.Setenv("HOME", t.TempDir())
			},
			wantErr: "does not exist",
		},
		{
			name:    "media path is not a directory",
			profile: "not-a-dir",
			setup: func(t *testing.T) {
				t.Helper()
				home := t.TempDir()
				t.Setenv("HOME", home)
				dir := filepath.Join(home, ".local", "share", "Anki2", "not-a-dir")
				if err := os.MkdirAll(dir, 0o750); err != nil {
					t.Fatalf("os.MkdirAll(%q): %v", dir, err)
				}
				fp := filepath.Join(dir, "collection.media")
				if err := os.WriteFile(fp, []byte("x"), 0o600); err != nil {
					t.Fatalf("os.WriteFile(%q): %v", fp, err)
				}
			},
			wantErr: "must be a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			runner, err := NewRunner(tt.profile)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("NewRunner(%q): got err=%v, want nil", tt.profile, err)
				}
				if runner == nil {
					t.Errorf("NewRunner(%q): got nil runner", tt.profile)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("NewRunner(%q): got err=%v, want substring %q", tt.profile, err, tt.wantErr)
			}
		})
	}
}

func TestRunner_postCSVRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		handler  http.HandlerFunc
		wantErr  string
	}{
		{
			name:     "invalid endpoint URL",
			endpoint: "http://example.com/%zz",
			wantErr:  "url.Parse",
		},
		{
			name:     "connection refused",
			endpoint: "http://127.0.0.1:1",
			wantErr:  "failed to connect to Anki",
		},
		{
			name: "non-200 with decodable message",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(map[string]string{"message": "bad deck"}); err != nil {
					t.Fatalf("json.NewEncoder().Encode(): %v", err)
				}
			},
			wantErr: `anki addon server returned message: "bad deck"`,
		},
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			endpoint := tt.endpoint
			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				t.Cleanup(server.Close)
				endpoint = server.URL
			}

			r := &Runner{client: http.DefaultClient, profile: "test-profile"}
			err := r.postCSVRequest(t.Context(), endpoint, "test-deck", strings.NewReader("data"))
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("postCSVRequest(%q): got err=%v, want nil", endpoint, err)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("postCSVRequest(%q): got err=%v, want substring=%q", endpoint, err, tt.wantErr)
			}
		})
	}
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(mockServer.Close)

	var mu sync.Mutex
	var saved []string
	saveFn := func(filename string, audio []byte) error {
		mu.Lock()
		defer mu.Unlock()
		saved = append(saved, filename)
		return nil
	}

	runner := &Runner{
		client:     mockServer.Client(),
		maxWorkers: 5,
		profile:    "test-profile",
		save:       saveFn,
	}

	raw, err := os.ReadFile("../testdata/anki-th-example.csv")
	if err != nil {
		t.Fatalf("os.ReadFile(): %v", err)
	}

	err = runner.Run(t.Context(), strings.NewReader(string(raw)), RunConfig{
		Speed:          "normal",
		Voice:          "th",
		Deck:           "test-deck",
		Endpoint:       mockServer.URL,
		Shuffle:        true,
		StripCSVHeader: true,
	})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(saved) == 0 {
		t.Error("Run(): save function to be called at least once")
	}
}

func TestRunner_Run_EmptyText(t *testing.T) {
	t.Parallel()

	runner := &Runner{
		client:     http.DefaultClient,
		maxWorkers: 1,
		save:       func(string, []byte) error { return nil },
	}

	input := "Text,HelperText,TextA,TextB,TextC,TextD\n\"{{c1::}}\",,,,,"
	err := runner.Run(t.Context(), strings.NewReader(input), RunConfig{Speed: "normal", Voice: "en"})
	if err == nil || !strings.Contains(err.Error(), "text is empty") {
		t.Errorf("Run(): got err=%v, want substring=%q", err, "text is empty")
	}
}
