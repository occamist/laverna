package yomitan

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAudioSourceListHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		https      bool
		wantAudio  bool
		wantScheme string
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
			name:       "valid request",
			query:      "term=Hello&reading=&language=en",
			wantAudio:  true,
			wantScheme: "http://",
		},
		{
			name:       "valid request behind https",
			query:      "term=Hello&reading=&language=en",
			https:      true,
			wantAudio:  true,
			wantScheme: "https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/?"+tt.query, nil)
			rec := httptest.NewRecorder()

			audioSourceListHandler(tt.https)(rec, req)

			resp := rec.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("audioSourceListHandler(): status=%d, want %d", resp.StatusCode, http.StatusOK)
			}
			if got := resp.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("audioSourceListHandler(): Content-Type=%q, want %q", got, "application/json")
			}

			var got audioSourceList
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatalf("json.Decode(): %v", err)
			}

			if got.Type != "audioSourceList" {
				t.Errorf("audioSourceListHandler(): type=%q, want %q", got.Type, "audioSourceList")
			}

			hasAudio := len(got.AudioSources) > 0
			if hasAudio != tt.wantAudio {
				t.Fatalf("audioSourceListHandler(): hasAudio=%v, want %v (audioSources=%v)", hasAudio, tt.wantAudio, got.AudioSources)
			}

			if hasAudio {
				gotURL := got.AudioSources[0].URL
				if !strings.HasPrefix(gotURL, tt.wantScheme) || !strings.Contains(gotURL, "/audio?") {
					t.Errorf("audioSourceListHandler(): url=%q, want scheme=%q pointing at /audio", gotURL, tt.wantScheme)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio?"+tt.query, nil)
			rec := httptest.NewRecorder()

			audioHandler(http.DefaultClient)(rec, req)

			resp := rec.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("audioHandler(): status=%d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantContent != "" {
				if got := resp.Header.Get("Content-Type"); got != tt.wantContent {
					t.Errorf("audioHandler(): Content-Type=%q, want %q", got, tt.wantContent)
				}
				if rec.Body.Len() == 0 {
					t.Errorf("audioHandler(): got empty body, want audio bytes")
				}
			}
		})
	}
}
