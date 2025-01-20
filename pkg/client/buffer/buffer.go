package buffer

import (
	"log"
	"sync"
	"time"

	"github.com/mmcdole/runes/pkg/client/ansi"
	"github.com/mmcdole/runes/pkg/client/types"
)

// Buffer handles storing and retrieving lines of text
type Buffer struct {
	mu          sync.RWMutex
	lines       []types.Line
	prompt      *types.Line
	processor   *ansi.Processor
	maxLines    int
}

// New creates a new buffer
func New(maxLines int) *Buffer {
	return &Buffer{
		lines:     make([]types.Line, 0),
		processor: ansi.NewProcessor(),
		maxLines:  maxLines,
	}
}

// Write handles both regular lines and prompts
func (b *Buffer) Write(line *types.Line) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Process ANSI codes
	line.Display = b.processor.Process(line.Raw)

	if line.IsPrompt {
		b.prompt = line
	} else {
		b.lines = append(b.lines, *line)
		if len(b.lines) > b.maxLines {
			b.lines = b.lines[1:]
		}
	}
}

// GetLines returns lines from start to end, including prompt if present at end
func (b *Buffer) GetLines(start, end int) []types.Line {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	// Validate range
	if start < 0 {
		start = 0
	}
	if end > len(b.lines) {
		end = len(b.lines)
	}
	if start >= end {
		return nil
	}
	
	// Get requested lines
	lines := make([]types.Line, end-start)
	copy(lines, b.lines[start:end])
	
	// If we're requesting up to the latest line and have a prompt,
	// include it
	if end == len(b.lines) && b.prompt != nil {
		lines = append(lines, *b.prompt)
	}
	
	return lines
}

// InputSent clears the current prompt
func (b *Buffer) InputSent() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prompt = nil
}

// Clear empties the buffer
func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = b.lines[:0]
	b.prompt = nil
	b.processor = ansi.NewProcessor()
	log.Printf("[Buffer] Cleared buffer")
}

// Len returns the number of complete lines in the buffer
func (b *Buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lines)
}

// HandlePrompt creates a prompt line from text
func (b *Buffer) HandlePrompt(text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.prompt = &types.Line{
		Raw:       text,
		Display:   text, // Will be processed by Write
		IsPrompt:  true,
		Timestamp: time.Now(),
	}
	
	// Process prompt like any other line
	b.Write(b.prompt)
}
