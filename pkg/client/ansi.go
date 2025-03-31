package client

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
	return s.Bold || s.Italic || s.Underline || s.Reverse || s.Blink ||
		s.FgColor != -1 || s.BgColor != -1
}

// ToANSI returns the ANSI escape codes for the current color state
func (s *ColorState) ToANSI() string {
	if !s.HasActiveAttributes() {
		return ""
	}

	var codes []string

	// Add attributes
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

	// Add colors
	if s.FgColor != -1 {
		codes = append(codes, fmt.Sprintf("38;5;%d", s.FgColor))
	}
	if s.BgColor != -1 {
		codes = append(codes, fmt.Sprintf("48;5;%d", s.BgColor))
	}

	return fmt.Sprintf("\x1b[%sm", strings.Join(codes, ";"))
}

// ProcessANSICode updates the state based on an ANSI code
func (s *ColorState) ProcessANSICode(code string) {
	// Extract the numeric codes
	parts := strings.Split(code[2:len(code)-1], ";")

	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "0": // Reset
			*s = *NewColorState()
		case "1": // Bold
			s.Bold = true
		case "3": // Italic
			s.Italic = true
		case "4": // Underline
			s.Underline = true
		case "5": // Blink
			s.Blink = true
		case "7": // Reverse
			s.Reverse = true
		case "22": // No Bold
			s.Bold = false
		case "23": // No Italic
			s.Italic = false
		case "24": // No Underline
			s.Underline = false
		case "25": // No Blink
			s.Blink = false
		case "27": // No Reverse
			s.Reverse = false
		case "38": // Foreground Color
			if i+2 < len(parts) && parts[i+1] == "5" {
				s.FgColor = atoi(parts[i+2])
				i += 2
			}
		case "48": // Background Color
			if i+2 < len(parts) && parts[i+1] == "5" {
				s.BgColor = atoi(parts[i+2])
				i += 2
			}
		}
	}
}

// Helper function to convert string to int
func atoi(s string) int {
	var n int
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}

// Processor handles ANSI escape sequence processing
type Processor struct {
	CurrentState *ColorState // Tracks unterminated color state from previous line
}

// NewProcessor creates a new ANSI processor
func NewProcessor() *Processor {
	return &Processor{
		CurrentState: NewColorState(),
	}
}

// Process processes ANSI escape sequences in a string
// It will:
// 1. Start with previous line's color state
// 2. Process all ANSI sequences in the input
// 3. Update the final state for the next line
// 4. Return the processed string
func (p *Processor) Process(input string) string {
	var result strings.Builder
	currentState := p.CurrentState.Clone()

	// Process the input
	content := input
	for len(content) > 0 {
		if strings.HasPrefix(content, "\x1b[") {
			// Find the end of ANSI sequence
			end := strings.IndexByte(content[2:], 'm')
			if end == -1 {
				// No end found, copy as-is
				result.WriteByte(content[0])
				content = content[1:]
				continue
			}
			end += 2 // Adjust for skipped prefix

			// Extract and process the sequence
			sequence := content[:end+1]
			currentState.ProcessANSICode(sequence)
			result.WriteString(sequence)
			content = content[end+1:]
		} else {
			result.WriteByte(content[0])
			content = content[1:]
		}
	}

	// Update final state
	p.CurrentState = currentState

	return result.String()
}
