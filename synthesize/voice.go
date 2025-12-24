package synthesize

import "slices"

// Voice represents ISO-639 language codes
// Taken from https://cloud.google.com/translate/docs/languages
// es-MX, es-ES or en-US, en-UK, en-AU voices are converted to regions and no longer available on web version, they are using different domains rather than language codes.
type Voice string

const (
	AfrikaansVoice           Voice = "af"
	AlbanianVoice            Voice = "sq"
	AmharicVoice             Voice = "am"
	ArabicVoice              Voice = "ar"
	BengaliVoice             Voice = "bn"
	BosnianVoice             Voice = "bs"
	BulgarianVoice           Voice = "bg"
	CantoneseVoice           Voice = "yue"
	CatalanVoice             Voice = "ca"
	ChineseSimplifiedVoice   Voice = "zh"
	ChineseTraditionalVoice  Voice = "zh-TW"
	CroatianVoice            Voice = "hr"
	CzechVoice               Voice = "cs"
	DanishVoice              Voice = "da"
	DutchVoice               Voice = "nl"
	EnglishVoice             Voice = "en"
	EstonianVoice            Voice = "et"
	FilipinoVoice            Voice = "tl"
	FinnishVoice             Voice = "fi"
	FrenchVoice              Voice = "fr"
	FrenchCanadianVoice      Voice = "fr-CA"
	GalicianVoice            Voice = "gl"
	GermanVoice              Voice = "de"
	GreekVoice               Voice = "el"
	GujaratiVoice            Voice = "gu"
	HausaVoice               Voice = "ha"
	HebrewVoice              Voice = "iw"
	HindiVoice               Voice = "hi"
	HungarianVoice           Voice = "hu"
	IcelandicVoice           Voice = "is"
	IndonesianVoice          Voice = "id"
	ItalianVoice             Voice = "it"
	JapaneseVoice            Voice = "ja"
	JavaneseVoice            Voice = "jw"
	KhmerVoice               Voice = "km"
	KoreanVoice              Voice = "ko"
	LatinVoice               Voice = "la"
	LatvianVoice             Voice = "lv"
	LithuanianVoice          Voice = "lt"
	MalayVoice               Voice = "ms"
	MalayalamVoice           Voice = "ml"
	MarathiVoice             Voice = "mr"
	MyanmarVoice             Voice = "my"
	NepaliVoice              Voice = "ne"
	NorwegianVoice           Voice = "no"
	PolishVoice              Voice = "pl"
	PortugueseBrazilianVoice Voice = "pt"
	PortugueseVoice          Voice = "pt-PT"
	PunjabiVoice             Voice = "pa"
	RomanianVoice            Voice = "ro"
	RussianVoice             Voice = "ru"
	SerbianVoice             Voice = "sr"
	SinhalaVoice             Voice = "si"
	SlovakVoice              Voice = "sk"
	SpanishVoice             Voice = "es"
	SundaneseVoice           Voice = "su"
	SwahiliVoice             Voice = "sw"
	SwedishVoice             Voice = "sv"
	TamilVoice               Voice = "ta"
	TeluguVoice              Voice = "te"
	ThaiVoice                Voice = "th"
	UkrainianVoice           Voice = "uk"
	UrduVoice                Voice = "ur"
	VietnameseVoice          Voice = "vi"
	WelshVoice               Voice = "cy"
)

// AllVoices returns a slice of all valid voice codes
var AllVoices = []string{
	string(AfrikaansVoice),
	string(AlbanianVoice),
	string(AmharicVoice),
	string(ArabicVoice),
	string(BengaliVoice),
	string(BosnianVoice),
	string(BulgarianVoice),
	string(CantoneseVoice),
	string(CatalanVoice),
	string(ChineseSimplifiedVoice),
	string(ChineseTraditionalVoice),
	string(CroatianVoice),
	string(CzechVoice),
	string(DanishVoice),
	string(DutchVoice),
	string(EnglishVoice),
	string(EstonianVoice),
	string(FilipinoVoice),
	string(FinnishVoice),
	string(FrenchVoice),
	string(FrenchCanadianVoice),
	string(GalicianVoice),
	string(GermanVoice),
	string(GreekVoice),
	string(GujaratiVoice),
	string(HausaVoice),
	string(HebrewVoice),
	string(HindiVoice),
	string(HungarianVoice),
	string(IcelandicVoice),
	string(IndonesianVoice),
	string(ItalianVoice),
	string(JapaneseVoice),
	string(JavaneseVoice),
	string(KhmerVoice),
	string(KoreanVoice),
	string(LatinVoice),
	string(LatvianVoice),
	string(LithuanianVoice),
	string(MalayVoice),
	string(MalayalamVoice),
	string(MarathiVoice),
	string(MyanmarVoice),
	string(NepaliVoice),
	string(NorwegianVoice),
	string(PolishVoice),
	string(PortugueseBrazilianVoice),
	string(PortugueseVoice),
	string(PunjabiVoice),
	string(RomanianVoice),
	string(RussianVoice),
	string(SerbianVoice),
	string(SinhalaVoice),
	string(SlovakVoice),
	string(SpanishVoice),
	string(SundaneseVoice),
	string(SwahiliVoice),
	string(SwedishVoice),
	string(TamilVoice),
	string(TeluguVoice),
	string(ThaiVoice),
	string(UkrainianVoice),
	string(UrduVoice),
	string(VietnameseVoice),
	string(WelshVoice),
}

// IsVoice validates if given string is a valid voice
func IsVoice(v string) bool {
	return slices.Contains(AllVoices, v)
}
