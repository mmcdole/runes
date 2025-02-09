package types

import "time"

// Line represents a line of text from the MUD or client
type Line struct {
    Content     string    // Processed content
    Raw         string    // Raw ANSI content
    Display     string    // Display content (processed ANSI)
    Flags       LineFlags // Processing flags
    Replacement *string   // Optional replacement content
    Timestamp   time.Time // When this line was received
}

// LineFlags contains processing flags for a line
type LineFlags struct {
    Gag         bool   // Don't display this line
    Matched     bool   // Line was matched by a trigger
    IsPrompt    bool   // Line is a prompt
    Complete    bool   // Line is complete (has newline)
    SkipLog     bool   // Don't log this line
    Source      string // Source of the line (mud, script, client)
}

// NewLine creates a new line with the given content
func NewLine(content string) *Line {
    return &Line{
        Content: content,
        Raw:     content,
        Display: content,
        Flags: LineFlags{
            Source: "mud",
        },
        Timestamp: time.Now(),
    }
}

// NewPrompt creates a new prompt line
func NewPrompt(content string) *Line {
    line := NewLine(content)
    line.Flags.IsPrompt = true
    return line
}

// NewScriptLine creates a new line from a script
func NewScriptLine(content string) *Line {
    line := NewLine(content)
    line.Flags.Source = "script"
    return line
}

// NewClientLine creates a new line from the client
func NewClientLine(content string) *Line {
    line := NewLine(content)
    line.Flags.Source = "client"
    return line
}

// Replace replaces the content of the line
func (l *Line) Replace(content string) {
    l.Replacement = &content
}

// GetContent returns the effective content of the line
func (l *Line) GetContent() string {
    if l.Replacement != nil {
        return *l.Replacement
    }
    return l.Content
}

// Clone creates a deep copy of the line
func (l *Line) Clone() *Line {
    clone := &Line{
        Content: l.Content,
        Raw:     l.Raw,
        Display: l.Display,
        Flags:   l.Flags,
        Timestamp: l.Timestamp,
    }
    if l.Replacement != nil {
        content := *l.Replacement
        clone.Replacement = &content
    }
    return clone
}
