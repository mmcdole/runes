package ansi

import (
	"strings"
)

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
	
	// Save state for next line
	p.CurrentState = currentState

	return result.String()
}
