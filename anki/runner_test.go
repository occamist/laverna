package anki

import (
	"errors"
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
