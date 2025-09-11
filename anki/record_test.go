package anki

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRecord_CleanedText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "cloze deletion at the end",
			text: "The capital of France is {{c1::Paris}}.",
			want: "The capital of France is Paris.",
		},
		{
			name: "cloze deletion at beginning",
			text: "{{c1::Hello}} world!",
			want: "Hello world!",
		},
		{
			name: "empty cloze deletion",
			text: "This is {{c1::}} empty.",
			want: "This is empty.",
		},
		{
			name: "cloze deletion with spaces",
			text: "The {{c1::quick brown fox}} jumps.",
			want: "The quick brown fox jumps.",
		},
		{
			name: "no cloze deletion marker",
			text: "This is plain text.",
		},
		{
			name: "malformed - no opening marker",
			text: "This is Paris}}.",
		},
		{
			name: "malformed - no closing marker",
			text: "This is {{c1::Paris.",
		},
		{
			name: "empty text",
		},
		{
			name: "only markers",
			text: "{{c1::}}",
		},
		{
			name: "multiple closing markers - takes first",
			text: "The {{c1::answer}} is }} here.",
		},
		{
			name: "nested braces in content",
			text: "Code: {{c1::func() { return }}}",
			want: "Code: func() { return }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Record{Text: tt.text}
			got := r.CleanedText()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CleanedText() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
