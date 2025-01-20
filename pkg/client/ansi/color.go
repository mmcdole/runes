package ansi

import (
	"fmt"
	"strings"
)

// ColorState tracks the current ANSI color and attribute state
type ColorState struct {
	Bold      bool
	Italic    bool
	Underline bool
	Reverse   bool
	Blink     bool
	FgColor   int // -1 for default, 0-255 for colors
	BgColor   int // -1 for default, 0-255 for colors
}

// NewColorState creates a new color state with default values
func NewColorState() *ColorState {
	return &ColorState{
		FgColor: -1,
		BgColor: -1,
	}
}

// Clone creates a copy of the current color state
func (s *ColorState) Clone() *ColorState {
	if s == nil {
		return NewColorState()
	}
	return &ColorState{
		Bold:      s.Bold,
		Italic:    s.Italic,
		Underline: s.Underline,
		Reverse:   s.Reverse,
		Blink:     s.Blink,
		FgColor:   s.FgColor,
		BgColor:   s.BgColor,
	}
}

// HasActiveAttributes returns true if any color or attribute is set
func (s *ColorState) HasActiveAttributes() bool {
	return s.Bold || s.Italic || s.Underline || s.Reverse || s.Blink || s.FgColor >= 0 || s.BgColor >= 0
}

// ToANSI returns the ANSI escape codes for the current color state
func (s *ColorState) ToANSI() string {
	var codes []string

	// Only add codes if we have any state to write
	if s.FgColor >= 0 {
		codes = append(codes, fmt.Sprintf("%d", s.FgColor+30))
	}
	if s.BgColor >= 0 {
		codes = append(codes, fmt.Sprintf("%d", s.BgColor+40))
	}
	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Underline {
		codes = append(codes, "4")
	}
	if s.Blink {
		codes = append(codes, "5")
	}
	if s.Reverse {
		codes = append(codes, "7")
	}

	if len(codes) == 0 {
		return ""
	}

	// Combine all codes into a single ANSI sequence
	return fmt.Sprintf("\033[%sm", strings.Join(codes, ";"))
}

// ProcessANSICode updates the state based on an ANSI code
func (s *ColorState) ProcessANSICode(code string) {
	if !strings.HasPrefix(code, "\033[") || !strings.HasSuffix(code, "m") {
		return
	}

	// Strip ESC[ and m
	code = code[2 : len(code)-1]
	
	// Handle empty code as reset
	if code == "" {
		*s = *NewColorState()
		return
	}

	// Process each code part
	for _, part := range strings.Split(code, ";") {
		num := 0
		fmt.Sscanf(part, "%d", &num)

		switch num {
		case 0: // Reset all
			*s = *NewColorState()
		case 1: // Bold
			s.Bold = true
		case 2: // Dim (we treat this as not bold)
			s.Bold = false
		case 3: // Italic
			s.Italic = true
		case 4: // Underline
			s.Underline = true
		case 5: // Blink
			s.Blink = true
		case 7: // Reverse
			s.Reverse = true
		case 21, 22: // Reset bold
			s.Bold = false
		case 23: // Reset italic
			s.Italic = false
		case 24: // Reset underline
			s.Underline = false
		case 25: // Reset blink
			s.Blink = false
		case 27: // Reset reverse
			s.Reverse = false
		case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground colors
			s.FgColor = num - 30
		case 39: // Default foreground
			s.FgColor = -1
		case 40, 41, 42, 43, 44, 45, 46, 47: // Background colors
			s.BgColor = num - 40
		case 49: // Default background
			s.BgColor = -1
		case 90, 91, 92, 93, 94, 95, 96, 97: // Bright foreground colors
			s.FgColor = num - 90 + 8
		case 100, 101, 102, 103, 104, 105, 106, 107: // Bright background colors
			s.BgColor = num - 100 + 8
		}
	}
}
