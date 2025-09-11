package anki

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
)

// Record is how anki CSV file should look like
type Record struct {
	Text                           string
	HelperText                     string
	TextA, TextB, TextC, TextD     string
	AudioA, AudioB, AudioC, AudioD string
	AudioAnswer                    string
}

// Row returns Record in correct order
func (r Record) Row() []string {
	return []string{
		r.Text,
		r.HelperText,
		r.TextA,
		r.TextB,
		r.TextC,
		r.TextD,
		r.AudioA,
		r.AudioB,
		r.AudioC,
		r.AudioD,
		r.AudioAnswer,
	}
}

// RandomizedRow returns Record with shuffled order of TextA, TextB, TextC, TextD
func (r Record) RandomizedRow() []string {
	type pair struct {
		text  string
		audio string
	}
	pairs := []pair{
		{r.TextA, r.AudioA},
		{r.TextB, r.AudioB},
		{r.TextC, r.AudioC},
		{r.TextD, r.AudioD},
	}

	rand.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	return []string{
		r.Text,
		r.HelperText,
		pairs[0].text,
		pairs[1].text,
		pairs[2].text,
		pairs[3].text,
		pairs[0].audio,
		pairs[1].audio,
		pairs[2].audio,
		pairs[3].audio,
		r.AudioAnswer,
	}
}

const (
	startCloze = "{{c1::"
	endCloze   = "}}"
)

// CleanedText returns the cleaned text without "{{c1::}}"
func (r Record) CleanedText() string {
	start := strings.Index(r.Text, startCloze)
	if start == -1 {
		return ""
	}

	end := strings.Index(r.Text[start:], endCloze)
	if end == -1 {
		return ""
	}

	// set to absolute end
	end = start + end
	answer := r.Text[start+len(startCloze) : end]

	cleanedText := r.Text[:start] + answer + r.Text[end+len(endCloze):]
	cleanedText = strings.ReplaceAll(cleanedText, "  ", " ")
	if strings.Contains(cleanedText, startCloze) || strings.Contains(cleanedText, endCloze) {
		return ""
	}
	return cleanedText
}

const inputHeaderSize = 6 // only won't include audio fields

// ReadCSVRecords reads records from anki CSV file
func ReadCSVRecords(r io.Reader) ([]Record, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = inputHeaderSize

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("%T.Read(): %w", reader, err)
	}
	h := []string{"Text", "HelperText", "TextA", "TextB", "TextC", "TextD"}
	for i := range len(header) {
		if header[i] != h[i] {
			return nil, fmt.Errorf("invalid header(%q) must be %q", header[i], h[i])
		}
	}

	var records []Record
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%T.Read(): %w", reader, err)
		}

		records = append(records, Record{
			Text:       record[0],
			HelperText: record[1],
			TextA:      record[2],
			TextB:      record[3],
			TextC:      record[4],
			TextD:      record[5],
		})
	}
	return records, nil
}

// WriteCSVRecords writes records into anki CSV file
func WriteCSVRecords(w io.Writer, records []Record, stripCSVHeader, shuffle bool) error {
	writer := csv.NewWriter(w)

	if !stripCSVHeader {
		h := []string{
			"Text", "HelperText", "TextA", "TextB", "TextC", "TextD",
			"AudioA", "AudioB", "AudioC", "AudioD", "AudioAnswer",
		}
		if err := writer.Write(h); err != nil {
			return fmt.Errorf("%T.Write(): %w", writer, err)
		}
	}

	for _, r := range records {
		var row []string
		if shuffle {
			row = r.RandomizedRow()
		} else {
			row = r.Row()
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("%T.Write(): %w", writer, err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("%T.Error(): %w", writer, err)
	}
	return nil
}
