package buffer

import (
    "sync"

    "github.com/mmcdole/runes/pkg/types"
)

// Buffer represents a scrollable buffer of lines
type Buffer struct {
    lines     []*types.Line
    maxLines  int
    mu        sync.RWMutex
}

// New creates a new buffer with the specified maximum lines
func New(maxLines int) *Buffer {
    return &Buffer{
        lines:    make([]*types.Line, 0, maxLines),
        maxLines: maxLines,
    }
}

// Add adds a line to the buffer
func (b *Buffer) Add(line *types.Line) {
    b.mu.Lock()
    defer b.mu.Unlock()

    // If we're at max capacity, remove oldest line
    if len(b.lines) >= b.maxLines {
        b.lines = b.lines[1:]
    }

    // Add new line
    b.lines = append(b.lines, line)
}

// GetLines returns a slice of lines from the buffer
func (b *Buffer) GetLines(start, count int) []*types.Line {
    b.mu.RLock()
    defer b.mu.RUnlock()

    if start < 0 {
        start = 0
    }
    if start >= len(b.lines) {
        return nil
    }

    end := start + count
    if end > len(b.lines) {
        end = len(b.lines)
    }

    return b.lines[start:end]
}

// GetPromptLines returns lines that are prompts
func (b *Buffer) GetPromptLines(count int) []*types.Line {
    b.mu.RLock()
    defer b.mu.RUnlock()

    var prompts []*types.Line
    for i := len(b.lines) - 1; i >= 0 && len(prompts) < count; i-- {
        if b.lines[i].Flags.IsPrompt {
            prompts = append(prompts, b.lines[i])
        }
    }
    return prompts
}

// Clear removes all lines from the buffer
func (b *Buffer) Clear() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.lines = b.lines[:0]
}

// Len returns the number of lines in the buffer
func (b *Buffer) Len() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return len(b.lines)
}

// AddPrompt adds a prompt line to the buffer
func (b *Buffer) AddPrompt(text string) {
    line := &types.Line{
        Raw:     text,
        Display: text,
        Flags: types.LineFlags{
            IsPrompt: true,
        },
    }
    b.Add(line)
}
