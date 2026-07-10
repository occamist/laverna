package synthesize

import "testing"

func TestNewSpeed(t *testing.T) {
	tests := []struct {
		str  string
		want Speed
	}{
		{
			str:  "normal",
			want: NormalSpeed,
		},
		{
			str:  "slower",
			want: SlowerSpeed,
		},
		{
			str:  "slowest",
			want: SlowestSpeed,
		},
		{
			str:  "bogus",
			want: NormalSpeed,
		},
		{
			str:  "",
			want: NormalSpeed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := NewSpeed(tt.str); got != tt.want {
				t.Errorf("NewSpeed(%v): got = %v, want = %v", tt.str, got, tt.want)
			}
		})
	}
}

func TestIsSpeed(t *testing.T) {
	tests := []struct {
		str  string
		want bool
	}{
		{str: "normal", want: true},
		{str: "slower", want: true},
		{str: "slowest", want: true},
		{str: "bogus", want: false},
		{str: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := IsSpeed(tt.str); got != tt.want {
				t.Errorf("IsSpeed(%v): got = %v, want = %v", tt.str, got, tt.want)
			}
		})
	}
}
