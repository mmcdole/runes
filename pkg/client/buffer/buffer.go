package buffer

import (
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/runes/pkg/client/ansi"
)

// Line represents a single line in the buffer
type Line struct {
	Content   string    // Processed content with ANSI codes
	Raw       string    // Original unprocessed content
	State     string    // ANSI state at start of line
	Timestamp time.Time // When the line was created
}

// Buffer stores and manages lines of text
type Buffer struct {
	lines      []Line
	processor  *ansi.LineProcessor
	maxLines   int
	mu         sync.RWMutex
}

// New creates a new buffer
func New() *Buffer {
	return &Buffer{
		lines:     make([]Line, 0),
		processor: ansi.NewLineProcessor(),
		maxLines:  10000, // Store 10k lines max
	}
}

// Write adds new content to the buffer
func (b *Buffer) Write(content string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	
	// Split content into lines
	rawLines := strings.Split(content, "\n")
	
	// Process each line
	for _, rawLine := range rawLines {
		// Skip empty lines that result from splitting on newlines
		if rawLine == "" {
			continue
		}
		
		// Process line through ANSI processor
		line := b.processor.ProcessLine(rawLine)
		
		// Store line with all its metadata
		b.lines = append(b.lines, Line{
			Content:   line.Content,
			Raw:       line.Raw,
			State:     b.processor.GetCurrentState(),
			Timestamp: time.Now(),
		})
	}
	
	// Trim if exceeding max lines
	if len(b.lines) > b.maxLines {
		excess := len(b.lines) - b.maxLines
		// Reset processor state to the state at the first kept line
		b.processor.Reset()
		if len(b.lines) > excess {
			b.processor.SetState(b.lines[excess].State)
		}
		b.lines = b.lines[excess:]
	}
}

// GetLines returns a slice of lines from the buffer
func (b *Buffer) GetLines(start, count int) []Line {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	if start < 0 {
		start = 0
	}
	if start >= len(b.lines) {
		return []Line{}
	}
	
	end := start + count
	if end > len(b.lines) {
		end = len(b.lines)
	}
	
	return b.lines[start:end]
}

// Len returns the number of lines in the buffer
func (b *Buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lines)
}

// Clear empties the buffer
func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = b.lines[:0]
	b.processor.Reset()
}
