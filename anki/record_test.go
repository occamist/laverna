package anki

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestReadCSVRecords(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Record
		wantErr error
	}{
		{
			name: "valid CSV with single record",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
The capital of {{c1::France}} is Paris,Helper text,A text,B text,C text,D text`,
			want: []Record{
				{
					Text:       "The capital of {{c1::France}} is Paris",
					HelperText: "Helper text",
					TextA:      "A text",
					TextB:      "B text",
					TextC:      "C text",
					TextD:      "D text",
				},
			},
		},
		{
			name: "valid CSV with multiple records",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
First {{c1::question}},First helper,A1,B1,C1,D1
Second {{c1::question}},Second helper,A2,B2,C2,D2`,
			want: []Record{
				{
					Text:       "First {{c1::question}}",
					HelperText: "First helper",
					TextA:      "A1",
					TextB:      "B1",
					TextC:      "C1",
					TextD:      "D1",
				},
				{
					Text:       "Second {{c1::question}}",
					HelperText: "Second helper",
					TextA:      "A2",
					TextB:      "B2",
					TextC:      "C2",
					TextD:      "D2",
				},
			},
		},
		{
			name: "valid CSV with empty fields and have commas in quoted field",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
"Question a,b,c {{c1::text}}",,,,,`,
			want: []Record{
				{
					Text:       "Question a,b,c {{c1::text}}",
					HelperText: "",
					TextA:      "",
					TextB:      "",
					TextC:      "",
					TextD:      "",
				},
			},
		},
		{
			name: "valid CSV with quoted fields containing commas",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
"Question with {{c1::comma, inside}}","Helper, with comma",A,B,C,D`,
			want: []Record{
				{
					Text:       "Question with {{c1::comma, inside}}",
					HelperText: "Helper, with comma",
					TextA:      "A",
					TextB:      "B",
					TextC:      "C",
					TextD:      "D",
				},
			},
		},
		{
			name: "valid CSV with leading spaces",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
  Question {{c1::text}},  Helper text,  A,  B,  C,  D`,
			want: []Record{
				{
					Text:       "Question {{c1::text}}",
					HelperText: "Helper text",
					TextA:      "A",
					TextB:      "B",
					TextC:      "C",
					TextD:      "D",
				},
			},
		},
		{
			name:  "header only CSV with no records",
			input: `Text,HelperText,TextA,TextB,TextC,TextD`,
			want:  nil,
		},
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: io.EOF,
		},
		{
			name: "invalid CSV with additional commas in unquoted field",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
Question a,b,c {{c1::text}},,,,,`,
			want:    nil,
			wantErr: errors.New("*csv.Reader.Read(): record on line 2: wrong number of fields"),
		},
		{
			name: "invalid CSV with no cloze deletion format",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
"Question a,b,c",,,,,`,
			want:    nil,
			wantErr: errors.New(`text("Question a,b,c") must contain cloze deletion format with {{c1::WORD}}`),
		},
		{
			name: "invalid header and wrong first column",
			input: `WrongText,HelperText,TextA,TextB,TextC,TextD
Question,Helper,A,B,C,D`,
			want:    nil,
			wantErr: errors.New(`invalid header("WrongText") must be "Text"`),
		},
		{
			name: "invalid header and wrong middle column",
			input: `Text,WrongHelper,TextA,TextB,TextC,TextD
Question,Helper,A,B,C,D`,
			want:    nil,
			wantErr: errors.New(`invalid header("WrongHelper") must be "HelperText"`),
		},
		{
			name: "invalid header and wrong last column",
			input: `Text,HelperText,TextA,TextB,TextC,WrongD
Question,Helper,A,B,C,D`,
			want:    nil,
			wantErr: errors.New(`invalid header("WrongD") must be "TextD"`),
		},
		{
			name: "too few fields in header",
			input: `Text,HelperText,TextA,TextB,TextC
Question,Helper,A,B,C`,
			want:    nil,
			wantErr: errors.New("*csv.Reader.Read(): record on line 1: wrong number of fields"),
		},
		{
			name: "too many fields in header",
			input: `Text,HelperText,TextA,TextB,TextC,TextD,ExtraColumn
Question,Helper,A,B,C,D,Extra`,
			want:    nil,
			wantErr: errors.New("*csv.Reader.Read(): record on line 1: wrong number of fields"),
		},
		{
			name: "too few fields in data record",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
Question,Helper,A,B,C`,
			want:    nil,
			wantErr: errors.New("*csv.Reader.Read(): record on line 2: wrong number of fields"),
		},
		{
			name: "too many fields in data record",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
Question,Helper,A,B,C,D,Extra`,
			want:    nil,
			wantErr: errors.New("*csv.Reader.Read(): record on line 2: wrong number of fields"),
		},
		{
			name: "malformed CSV and unclosed quote",
			input: `Text,HelperText,TextA,TextB,TextC,TextD
"Question with unclosed quote,Helper,A,B,C,D`,
			want:    nil,
			wantErr: errors.New(`*csv.Reader.Read(): parse error on line 2, column 45: extraneous or missing " in quoted-field`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := ReadCSVRecords(reader)

			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if !cmp.Equal(tt.wantErr.Error(), err.Error()) {
					t.Errorf("ReadCSVRecords(): wantErr=%v gotError=%v", tt.wantErr, err)
				}
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ReadCSVRecords() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteCSVRecords(t *testing.T) {
	const f = "../testdata/Aout.csv"
	raw, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", f, err)
	}

	tests := []struct {
		name           string
		records        []Record
		stripCSVHeader bool
		shuffle        bool
		want           string
		wantErr        error
	}{
		{
			name: "write with records with header",
			records: []Record{
				{
					Text:        "ฉันชอบ{{c1::ฟัง}}เพลง",
					HelperText:  "I like to listen to music",
					TextA:       "ฟัง",
					TextB:       "เล่น",
					TextC:       "ดู",
					TextD:       "อ่าน",
					AudioA:      "[sound:a.mp3]",
					AudioB:      "[sound:b.mp3]",
					AudioC:      "[sound:c.mp3]",
					AudioD:      "[sound:d.mp3]",
					AudioAnswer: "[sound:e.mp3]",
				},
			},
			stripCSVHeader: false,
			want:           string(raw),
		},
		{
			name:           "write empty records without header",
			records:        []Record{},
			stripCSVHeader: true,
			want:           "",
		},
		{
			name:           "write empty records with header",
			records:        []Record{},
			stripCSVHeader: false,
			want:           "Text,HelperText,TextA,TextB,TextC,TextD,AudioA,AudioB,AudioC,AudioD,AudioAnswer\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteCSVRecords(&buf, tt.records, tt.stripCSVHeader, tt.shuffle)

			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				if !cmp.Equal(tt.wantErr.Error(), err.Error()) {
					t.Errorf("WriteCSVRecords(): wantErr=%v gotError=%v", tt.wantErr, err)
				}
			}

			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WriteCSVRecords() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
