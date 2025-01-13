package buffer

import (
	"strings"
	"sync"
)

type Buffer struct {
	lines    []string
	maxLines int
	mutex    sync.RWMutex
}

type BufferConfig struct {
	MaxLines   int
	InitialCap int
}

// NewBuffer creates a new Buffer with the given configuration
func NewBuffer(config BufferConfig) *Buffer {
	if config.MaxLines <= 0 {
		config.MaxLines = 1000 // reasonable default
	}
	if config.InitialCap <= 0 {
		config.InitialCap = 100
	}
	return &Buffer{
		lines:    make([]string, 0, config.InitialCap),
		maxLines: config.MaxLines,
	}
}

// Append adds a new line to the buffer, maintaining the max size
func (b *Buffer) Append(line string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.lines) >= b.maxLines {
		copy(b.lines, b.lines[1:])
		b.lines[len(b.lines)-1] = line
	} else {
		b.lines = append(b.lines, line)
	}
}

// GetLines returns a slice of lines between start and end
func (b *Buffer) GetLines(start, end int) []string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if start < 0 {
		start = 0
	}
	if end > len(b.lines) {
		end = len(b.lines)
	}
	if start >= end {
		return nil
	}

	result := make([]string, end-start)
	copy(result, b.lines[start:end])
	return result
}

// Clear removes all lines from the buffer
func (b *Buffer) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.lines = b.lines[:0]
}

// Length returns the current number of lines in the buffer
func (b *Buffer) Length() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.lines)
}

// Content returns all lines joined with newlines - used for testing and debugging
func (b *Buffer) Content() string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return strings.Join(b.lines, "\n")
}

// BufferManager manages multiple named buffers
type BufferManager struct {
	buffers  map[string]*Buffer
	current  string
	mu       sync.RWMutex
	onUpdate func() // Callback when buffer is updated
}

// NewBufferManager creates a new BufferManager
func NewBufferManager() *BufferManager {
	bm := &BufferManager{
		buffers: make(map[string]*Buffer),
	}
	bm.current = "main"
	bm.buffers[bm.current] = NewBuffer(BufferConfig{MaxLines: 1000})
	return bm
}

// SetUpdateCallback sets the callback for buffer updates
func (bm *BufferManager) SetUpdateCallback(cb func()) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.onUpdate = cb
}

// GetCurrentBuffer returns the current buffer
func (bm *BufferManager) GetCurrentBuffer() *Buffer {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	buf, ok := bm.buffers[bm.current]
	if !ok {
		buf = NewBuffer(BufferConfig{MaxLines: 1000})
		bm.buffers[bm.current] = buf
	}
	return buf
}

// SetCurrentBuffer sets the current buffer
func (bm *BufferManager) SetCurrentBuffer(name string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.buffers[name]; !exists {
		bm.buffers[name] = NewBuffer(BufferConfig{MaxLines: 1000})
	}
	bm.current = name
}

// AddLine adds a line to the current buffer
func (bm *BufferManager) AddLine(line string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.current == "" {
		bm.current = "main"
	}

	buf, ok := bm.buffers[bm.current]
	if !ok {
		buf = NewBuffer(BufferConfig{MaxLines: 1000})
		bm.buffers[bm.current] = buf
	}

	// Split the input into lines and add each one separately
	lines := strings.Split(line, "\n")
	for _, l := range lines {
		if l != "" { // Skip empty lines
			buf.Append(l)
		}
	}

	// Notify UI of update
	if bm.onUpdate != nil {
		go bm.onUpdate() // Run callback in goroutine to avoid deadlock
	}
}

// LineProcessor processes incoming lines
type LineProcessor struct {
	mutex       sync.Mutex
	buffer      []byte
	inSequence  bool
}

// NewLineProcessor creates a new LineProcessor
func NewLineProcessor() *LineProcessor {
	return &LineProcessor{
		buffer: make([]byte, 0, 1024),
	}
}

// ProcessLine processes a line of text
func (lp *LineProcessor) ProcessLine(line string) string {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	// Return as-is to preserve ANSI sequences
	return line
}

// Write implements io.Writer for LineProcessor
func (lp *LineProcessor) Write(p []byte) (n int, err error) {
	// Process the line but preserve ANSI sequences
	_ = lp.ProcessLine(string(p))
	return len(p), nil
}
