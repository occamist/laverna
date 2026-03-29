// Package synthesize implements an algorithm for synthesizing text to audio.
package synthesize

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
)

// Opt consists of parameters for generating audio
type Opt struct {
	Speed Speed
	Voice Voice
	Text  string
}

// ErrEmptyCSV occurs when empty csv is given
var ErrEmptyCSV = errors.New("empty csv")

// UnmarshalCSV reads raw bytes from CSV and turns into Opts
func UnmarshalCSV(raw []byte) ([]Opt, error) {
	if len(raw) == 0 {
		return nil, ErrEmptyCSV
	}

	reader := csv.NewReader(bytes.NewReader(raw))
	reader.TrimLeadingSpace = true
	record, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("%T.Read(): %v", reader, err)
	}
	header := []string{"speed", "voice", "text"}
	if !slices.Equal(header, record) {
		return nil, fmt.Errorf("header record(%v) is not the correct header(%v)", record, header)
	}

	var opts []Opt
	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%T.Read(): %w", reader, err)
		}

		speed, voice, text := record[0], record[1], record[2]
		var opt Opt
		opt.Speed = NewSpeed(strings.ToLower(speed))
		opt.Voice = Voice(voice)
		if !strings.Contains(voice, "-") {
			opt.Voice = Voice(strings.ToLower(voice))
		}
		opt.Text = text
		opts = append(opts, opt)
	}
	return opts, nil
}

const rpcID = "jQ1olc"

// Request will look as below, since it is a form, the key is f.req
// and the URL encoded value is going to be
/*
	[
		[
			[
			"jQ1olc",
			"[
				\"สวัสดีชาวโลก วันนี้เราจะมาพูดคุยกันถึงปัญหาของโลก\",     // Text
				\"th\",                                // Voice
				null,
				null,
				[2]									   // Speed
			]",
			null,
			"generic"
			]
		]
	]
*/
func makeFormData(opt Opt) (string, error) {
	genericOpts := []any{opt.Text, opt.Voice, nil, nil, []Speed{opt.Speed}}
	rawOpts, err := json.Marshal(genericOpts)
	if err != nil {
		return "", fmt.Errorf("json.Marshal(%v): %v", genericOpts, err)
	}

	genericData := [][][]any{
		{
			{rpcID, string(rawOpts), nil, "generic"},
		},
	}
	rawData, err := json.Marshal(genericData)
	if err != nil {
		return "", fmt.Errorf("json.Marshal(%v): %v", genericData, err)
	}

	form := make(url.Values)
	form.Set("f.req", string(rawData))
	return form.Encode(), nil
}

// Response will look as below, this function parses base64 data to MP3 format
/*
	)]}'
	[
		["wrb.fr","jQ1olc","[\"<base 64 data>\"]", null, null, null, "generic"],
		["di", 208],
		["af.httprm",208,"6046482986355911791",35]
	]
*/
func parseAudio(raw []byte) ([]byte, error) {
	lines := strings.Split(string(raw), "\n")

	// Try to find the line that contains the array with base64 audio data
	var audioLine string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, numbers, or the ")]}'
		if line == "" || line == ")]}''" || (len(line) > 0 && line[0] >= '0' && line[0] <= '9') {
			continue
		}

		if strings.HasPrefix(line, "[") { // Found what looks like a JSON array
			var jsonArray [][]any
			if err := json.Unmarshal([]byte(line), &jsonArray); err != nil {
				continue
			}
			// Check if the array contains the audio data
			if len(jsonArray) > 0 && len(jsonArray[0]) > 2 && jsonArray[0][2] != nil {
				audioLine = line
				break
			}
		}
	}

	if audioLine == "" {
		return nil, errors.New("no audio line found")
	}

	var audioLineSubParts [][]any
	if err := json.Unmarshal([]byte(audioLine), &audioLineSubParts); err != nil {
		return nil, fmt.Errorf("json.Unmarshal(%v): %v", audioLine, err)
	}

	audioLineSubPart, ok := audioLineSubParts[0][2].(string)
	if !ok {
		return nil, errors.New("no audio line sub part found")
	}

	var base64EncodedAudio []string
	if err := json.Unmarshal([]byte(audioLineSubPart), &base64EncodedAudio); err != nil {
		return nil, fmt.Errorf("json.Unmarshal(%v): %v", audioLineSubPart, err)
	}

	if len(base64EncodedAudio) == 0 {
		return nil, errors.New("no base64 encoded audio found")
	}

	audio, err := base64.StdEncoding.DecodeString(base64EncodedAudio[0])
	if err != nil {
		return nil, fmt.Errorf("%T.DecodeString(%v): %v", base64.StdEncoding, base64EncodedAudio[0], err)
	}
	return audio, nil
}

const hostname = "https://translate.google.com"

// ErrTextTooLong occurs when given text is longer than 200 characters
var ErrTextTooLong = errors.New("text must be less than 200 chars")

// Run produces the audio with a http client and a given option
func Run(ctx context.Context, c *http.Client, opt Opt) (_ []byte, err error) {
	const URL = hostname + "/_/TranslateWebserverUi/data/batchexecute"

	if len(opt.Text) > 200 {
		return nil, ErrTextTooLong
	}
	if c == nil {
		c = http.DefaultClient
	}

	formData, err := makeFormData(opt)
	if err != nil {
		return nil, fmt.Errorf("makeFormData(%v): %v", opt, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewBufferString(formData))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext(): %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", hostname)
	req.Header.Set("Referer", hostname)
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%T.Do(): %w", c, err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("resp.Body.Close(): %v", closeErr)
		}
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll(): %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		if dump, err := httputil.DumpResponse(resp, true); err != nil {
			return nil, fmt.Errorf("%v returned: %v\n %v", URL, resp.Status, string(dump))
		}
		return nil, fmt.Errorf("%v returned: %v", URL, resp.Status)
	}
	return parseAudio(raw)
}
