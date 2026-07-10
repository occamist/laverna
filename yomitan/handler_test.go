package yomitan

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(NewHandler(http.DefaultClient, false))
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/?term=Hello&language=en") //nolint:noctx // test-only request against a server we just started
	if err != nil {
		t.Fatalf("http.Get(): %v", err)
	}
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /: got status=%d, want status=%d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("GET /: got content type=%q, want content type=%q", got, "application/json")
	}

	audioResp, err := http.Get(server.URL + "/audio?term=Hello&language=en") //nolint:noctx // test-only request against a server we just started
	if err != nil {
		t.Fatalf("http.Get(): %v", err)
	}
	t.Cleanup(func() {
		_ = audioResp.Body.Close()
	})
	if audioResp.StatusCode != http.StatusOK {
		t.Errorf("GET /audio: got status=%d, want status=%d", audioResp.StatusCode, http.StatusOK)
	}
	if got := audioResp.Header.Get("Content-Type"); got != "audio/mpeg" {
		t.Errorf("GET /audio: got content type=%q, want content type=%q", got, "audio/mpeg")
	}
}

// brokenResponseWriter always fails on Write()
type brokenResponseWriter struct {
	header http.Header
}

func (w *brokenResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *brokenResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("broken pipe")
}

func (w *brokenResponseWriter) WriteHeader(int) {}

func TestWriteAudioSourceList_WriteError(t *testing.T) {
	t.Parallel()
	writeAudioSourceList(&brokenResponseWriter{}, []audioSource{{Name: "Laverna", URL: "http://example.com/audio"}})
}

func TestAudioHandler_WriteError(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio?term=Hello&language=en", nil)
	handler := audioHandler(http.DefaultClient)
	handler(&brokenResponseWriter{}, req)
}

func TestAudioSourceListHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		query          string
		https          bool
		wantAudioCount int
		wantScheme     string
	}{
		{
			name: "missing all params",
		},
		{
			name:  "invalid language",
			query: "term=hello&reading=hello&language=XYZTESTXYZ",
		},
		{
			name:  "missing term and reading",
			query: "language=en",
		},
		{
			name:           "valid request",
			query:          "term=Hello&reading=&language=en",
			wantAudioCount: 1,
			wantScheme:     "http://",
		},
		{
			name:           "valid request behind https",
			query:          "term=Hello&reading=&language=en",
			https:          true,
			wantAudioCount: 1,
			wantScheme:     "https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/?"+tt.query, nil)
			w := httptest.NewRecorder()

			handler := audioSourceListHandler(tt.https)
			handler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("audioSourceListHandler(): got status=%d, want status=%d", resp.StatusCode, http.StatusOK)
			}
			if got := resp.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("audioSourceListHandler(): got content type=%q, want content type=%q", got, "application/json")
			}

			var got audioSourceList
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatalf("json.Decode(): %v", err)
			}

			if got.Type != "audioSourceList" {
				t.Errorf("audioSourceListHandler(): got type=%q, want type=%q", got.Type, "audioSourceList")
			}

			gotAudioCount := len(got.AudioSources)
			if gotAudioCount != tt.wantAudioCount {
				t.Fatalf("audioSourceListHandler(): got audio count=%v, want audio count=%v", gotAudioCount, tt.wantAudioCount)
			}

			if gotAudioCount > 1 {
				gotURL := got.AudioSources[0].URL
				if !strings.HasPrefix(gotURL, tt.wantScheme) || !strings.Contains(gotURL, "/audio?") {
					t.Errorf("audioSourceListHandler(): got url=%q, want scheme=%q pointing at /audio?", gotURL, tt.wantScheme)
				}
			}
		})
	}
}

func TestAudioHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		query       string
		wantStatus  int
		wantContent string
	}{
		{
			name:       "missing all params",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid language",
			query:      "term=hello&language=XYZTESTXYZ",
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "valid request",
			query:       "term=Hello&language=en",
			wantStatus:  http.StatusOK,
			wantContent: "audio/mpeg",
		},
		{
			name:       "synthesize error text must be less than 200 chars",
			query:      "term=" + strings.Repeat("a", 201) + "&language=en",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio?"+tt.query, nil)
			w := httptest.NewRecorder()

			handler := audioHandler(http.DefaultClient)
			handler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("audioHandler(): got status=%d, want status=%d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantContent != "" {
				if got := resp.Header.Get("Content-Type"); got != tt.wantContent {
					t.Errorf("audioHandler(): got content type=%q, want content type=%q", got, tt.wantContent)
				}
				if w.Body.Len() == 0 {
					t.Errorf("audioHandler(): got empty body, want audio bytes")
				}
			}
		})
	}
}
