package yomitan

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/occamist/laverna/synthesize"
)

// NewHandler returns an http.Handler implementing Yomitan's "Custom URL (JSON)" audio
// source. Point Yomitan at http://host:port?term={term}&reading={reading}&language={language}
//
// It exposes two routes: "/" answers Yomitan's discovery request with a JSON list of
// candidate audio URLs, and "/audio" serves the synthesized audio bytes for one of those
// URLs. Two routes are required because Yomitan re-fetches every audioSources[].url
// through its own network layer, which only supports plain http(s) URLs, not data URIs.
func NewHandler(client *http.Client, https bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", audioSourceListHandler(https))
	mux.HandleFunc("/audio", audioHandler(client))
	return mux
}

// synthesizeParams picks the text to synthesize (preferring reading over term,
// since that's the actual pronunciation) and validates the requested voice.
func synthesizeParams(term, reading, language string) (synthesize.Opt, error) {
	text := strings.TrimSpace(reading)
	if text == "" {
		text = strings.TrimSpace(term)
	}
	if text == "" {
		return synthesize.Opt{}, errors.New("term and reading must not both be empty")
	}
	if !synthesize.IsVoice(language) {
		return synthesize.Opt{}, fmt.Errorf("language(%q) is not a valid voice", language)
	}
	return synthesize.Opt{Text: text, Voice: synthesize.Voice(language), Speed: synthesize.NormalSpeed}, nil
}

// audioSource is a single entry of a Yomitan custom audio list response.
type audioSource struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url"`
}

// audioSourceList is the response body a Yomitan "Custom URL (JSON)" audio source expects.
type audioSourceList struct {
	Type         string        `json:"type"`
	AudioSources []audioSource `json:"audioSources"`
}

func writeAudioSourceList(w http.ResponseWriter, sources []audioSource) {
	w.Header().Set("Content-Type", "application/json")
	resp := audioSourceList{Type: "audioSourceList", AudioSources: sources}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("%T.Encode(): %v", json.NewEncoder(w), err)
	}
}

func audioSourceListHandler(https bool) http.HandlerFunc {
	scheme := "http"
	if https {
		scheme = "https"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		term := q.Get("term")
		reading := q.Get("reading")
		language := q.Get("language")
		log.Printf("language = %s, reading = %s, term = %s\n", language, reading, term) //nolint:gosec // query params logged for local debugging only

		if _, err := synthesizeParams(term, reading, language); err != nil { // means not found, should return empty list as 200
			writeAudioSourceList(w, []audioSource{})
			return
		}

		audioURL := url.URL{Scheme: scheme, Host: r.Host, Path: "/audio", RawQuery: q.Encode()}
		sources := []audioSource{{Name: "Laverna", URL: audioURL.String()}}
		writeAudioSourceList(w, sources)
	}
}

func audioHandler(client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		opt, err := synthesizeParams(q.Get("term"), q.Get("reading"), q.Get("language"))
		if err != nil {
			http.Error(w, "no audio available", http.StatusNotFound)
			return
		}

		audio, err := synthesize.Run(r.Context(), client, opt)
		if err != nil {
			log.Printf("synthesize.Run(): %v\n", err)
			http.Error(w, "failed to synthesize audio", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		if _, err := w.Write(audio); err != nil { //nolint:gosec // binary audio bytes with a non-HTML content type, not attacker-controlled markup
			log.Printf("%T.Write(): %v\n", w, err)
			http.Error(w, "failed to write audio as response", http.StatusInternalServerError)
		}
	}
}
