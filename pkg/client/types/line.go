package types

import "time"

// Line represents a line of text from the MUD
type Line struct {
    Raw     string    // Original bytes from network
    Display string    // ANSI processed for terminal display
    
    // Line state flags
    IsPrompt  bool      // Whether this is a prompt
    Complete  bool      // Whether this is a complete line
    Matched   bool      // Whether this line was matched by a trigger
    Gag       bool      // Whether this line should be hidden from display
    SkipLog   bool      // Whether this line should be excluded from logging
    Timestamp time.Time // When this line was received
}

// NewLine creates a new Line with raw content
func NewLine(raw string) *Line {
    return &Line{
        Raw:       raw,
        Timestamp: time.Now(),
    }
}

// NewPrompt creates a new prompt Line
func NewPrompt(raw string) *Line {
    return &Line{
        Raw:       raw,
        IsPrompt:  true,
        Complete:  false,
        Timestamp: time.Now(),
    }
}
