package ansi

import "time"

// Line represents a line of text with its associated color state
type Line struct {
	Raw       string    // The original unprocessed content
	Content   string    // The processed content including ANSI codes
	Timestamp time.Time // When the line was created
}

// LineProcessor processes lines of text, maintaining color state between lines
type LineProcessor struct {
	currentState *colorState // Tracks unterminated color state from previous line
}

// NewLineProcessor creates a new line processor
func NewLineProcessor() *LineProcessor {
	return &LineProcessor{
		currentState: newColorState(),
	}
}

// ProcessLine processes a line of text, tracking unterminated color states
func (p *LineProcessor) ProcessLine(content string) *Line {
	// Create line with timestamp and raw content
	line := &Line{
		Raw:       content,
		Timestamp: time.Now(),
	}

	// Track color state as we process the line
	state := p.currentState.clone() // Start with previous state
	wasReset := false

	// If we have unterminated color state from previous line, prepend it
	if state.hasColor() {
		line.Content = state.toANSI() + content
	} else {
		line.Content = content
	}

	for i := 0; i < len(content); i++ {
		if content[i] == '\033' && i+1 < len(content) && content[i+1] == '[' {
			// Find the end of the sequence
			end := -1
			for j := i + 2; j < len(content); j++ {
				if content[j] == 'm' {
					end = j
					break
				}
			}
			if end == -1 {
				continue
			}

			// Process this sequence
			seq := content[i+2 : end]
			if seq == "0" || seq == "" {
				state = newColorState() // Create fresh state after reset
				wasReset = true
			} else {
				if wasReset {
					state = newColorState() // Start fresh after a reset
					wasReset = false
				}
				state.processANSI(seq)
			}
			i = end // Skip to end of sequence
		}
	}

	// Store final state for next line
	p.currentState = state

	return line
}

// GetCurrentState returns the current ANSI state as a string
func (p *LineProcessor) GetCurrentState() string {
	if !p.currentState.hasColor() {
		return ""
	}
	return p.currentState.toANSI()
}

// SetState sets the current ANSI state from a string
func (p *LineProcessor) SetState(state string) {
	if state == "" {
		p.currentState = newColorState()
		return
	}
	
	p.currentState = newColorState()
	p.currentState.processANSI(state)
}

// Reset resets the processor's color state
func (p *LineProcessor) Reset() {
	p.currentState = newColorState()
}
