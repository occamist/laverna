package synthesize

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBatchRunner(t *testing.T) {
	const maxWorkers = 5
	saveErr := errors.New("save error")
	temp := t.TempDir()

	tests := []struct {
		name       string
		opts       []Opt
		saveFn     func(string, []byte) error
		wantErr    error
		wantAudios []string
		ctx        context.Context //nolint:containedctx // one test requires a cancelled context
	}{
		{
			name: "successful batch run",
			opts: []Opt{
				{Text: "test1", Voice: EnglishVoice},
				{Text: "test2", Voice: EnglishVoice},
			},
			saveFn: func(text string, audio []byte) error {
				return os.WriteFile(filepath.Join(temp, text+".mp3"), audio, 0o600)
			},
			wantAudios: []string{
				filepath.Join(temp, "test1.mp3"),
				filepath.Join(temp, "test2.mp3"),
			},
			ctx: t.Context(),
		},
		{
			name: "context cancelled",
			opts: []Opt{
				{Text: "test3", Voice: EnglishVoice},
			},
			saveFn:  func(string, []byte) error { return nil },
			wantErr: context.Canceled,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel() // Cancel immediately
				return ctx
			}(),
		},
		{
			name: "custom save function error",
			opts: []Opt{
				{Text: "test4", Voice: EnglishVoice},
			},
			saveFn: func(string, []byte) error {
				return saveErr
			},
			wantErr: saveErr,
			ctx:     t.Context(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewRunner(
				WithClient(&http.Client{}),
				WithMaxWorkers(maxWorkers),
				WithSaveFunc(tt.saveFn),
			)

			err := runner.Run(tt.ctx, tt.opts)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				t.Errorf("%T.Run(): wantErr=%v, gotErr=%v", runner, tt.wantErr, err)
			}

			for _, wantAudio := range tt.wantAudios {
				info, err := os.Stat(wantAudio)
				if os.IsNotExist(err) {
					t.Errorf("want file(%q) doesn't exist", wantAudio)
					continue
				}
				if info.Size() == 0 {
					t.Errorf("want file(%q) has no bytes", wantAudio)
				}
			}
		})
	}
}
