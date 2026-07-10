package synthesize

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var testVoices = []Voice{
	AfrikaansVoice,
	AlbanianVoice,
	AmharicVoice,
	ArabicVoice,
	BengaliVoice,
	BosnianVoice,
	BulgarianVoice,
	CantoneseVoice,
	CatalanVoice,
	ChineseSimplifiedVoice,
	ChineseTraditionalVoice,
	CroatianVoice,
	CzechVoice,
	DanishVoice,
	DutchVoice,
	EnglishVoice,
	EstonianVoice,
	FilipinoVoice,
	FinnishVoice,
	FrenchVoice,
	FrenchCanadianVoice,
	GalicianVoice,
	GermanVoice,
	GreekVoice,
	GujaratiVoice,
	HausaVoice,
	HebrewVoice,
	HindiVoice,
	HungarianVoice,
	IcelandicVoice,
	IndonesianVoice,
	ItalianVoice,
	JapaneseVoice,
	JavaneseVoice,
	KhmerVoice,
	KoreanVoice,
	LatinVoice,
	LatvianVoice,
	LithuanianVoice,
	MalayVoice,
	MalayalamVoice,
	MarathiVoice,
	MyanmarVoice,
	NepaliVoice,
	NorwegianVoice,
	PolishVoice,
	PortugueseBrazilianVoice,
	PortugueseVoice,
	PunjabiVoice,
	RomanianVoice,
	RussianVoice,
	SerbianVoice,
	SinhalaVoice,
	SlovakVoice,
	SpanishVoice,
	SundaneseVoice,
	SwahiliVoice,
	SwedishVoice,
	TamilVoice,
	TeluguVoice,
	ThaiVoice,
	UkrainianVoice,
	UrduVoice,
	VietnameseVoice,
	WelshVoice,
}

func TestIsVoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		str  string
		want bool
	}{
		{str: "en", want: true},
		{str: "th", want: true},
		{str: "zh-TW", want: true},
		{str: "bogus", want: false},
		{str: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			t.Parallel()
			if got := IsVoice(tt.str); got != tt.want {
				t.Errorf("IsVoice(%v): got = %v, want = %v", tt.str, got, tt.want)
			}
		})
	}
}

func TestRun_AllVoices(t *testing.T) {
	t.Parallel()

	type Test struct {
		name    string
		client  *http.Client
		opt     Opt
		wantErr error
	}
	tests := make([]Test, 0, len(testVoices))

	for _, voice := range testVoices {
		opt := Opt{
			Speed: NormalSpeed,
			Voice: voice,
			Text:  "hello",
		}
		tests = append(tests, Test{
			name:    "language voice:" + string(opt.Voice),
			client:  &http.Client{},
			opt:     opt,
			wantErr: nil,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			audio, err := Run(t.Context(), tt.client, tt.opt)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				t.Errorf("Run(%v): wantErr=%v, gotErr=%v", tt.opt, tt.wantErr, err)
			}

			if len(audio) < 1 {
				t.Errorf("Run(%v): audio len(%v) must be greater than 1", tt.opt, len(audio))
			}
		})
	}
}
