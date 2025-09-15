package synthesize

import (
	"slices"
)

// Speed is the pronunciation speed of the voice
type Speed int

const (
	// NormalSpeed is the fastest speed
	NormalSpeed Speed = iota
	// SlowerSpeed is the middle speed
	SlowerSpeed
	// SlowestSpeed is the slowest speed
	SlowestSpeed
)

// NewSpeed returns the speed
func NewSpeed(s string) Speed {
	switch s {
	case NormalSpeed.String():
		return NormalSpeed
	case SlowerSpeed.String():
		return SlowerSpeed
	case SlowestSpeed.String():
		return SlowestSpeed
	default:
		return NormalSpeed
	}
}

// String returns the string representation of speed
func (s Speed) String() string {
	return []string{"normal", "slower", "slowest"}[s]
}

// IsSpeed validates if given string is speed
func IsSpeed(s string) bool {
	return slices.Contains([]string{"normal", "slower", "slowest"}, s)
}
