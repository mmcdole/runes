package client

import (
	"fmt"
	"io"
	"sync"
)

const (
	MainBuffer = "main" // Default buffer for MUD output
)

type Display struct {
	output    io.Writer
	buffers   map[string]*Buffer
	current   string
	lineCount int
	mu        sync.RWMutex
}

func (d *Display) Write(p []byte) (n int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Fprintln(d.output, string(p))
	return len(p), nil
}

type Buffer struct {
	Name    string
	Lines   []string
	Visible bool
}

func NewDisplay(output io.Writer) *Display {
	d := &Display{
		output:    output,
		buffers:   make(map[string]*Buffer),
		current:   MainBuffer,
		lineCount: 50,
	}
	d.buffers[MainBuffer] = &Buffer{Name: MainBuffer, Visible: true}
	return d
}

func (d *Display) WriteText(text string, buffer string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if buffer == "" {
		buffer = d.current
	}

	if _, exists := d.buffers[buffer]; !exists {
		d.buffers[buffer] = &Buffer{Name: buffer, Visible: true}
	}

	d.buffers[buffer].Lines = append(d.buffers[buffer].Lines, text)
	if buffer == d.current {
		fmt.Fprintln(d.output, text)
	}
}

func (d *Display) SwitchBuffer(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.buffers[name]; !exists {
		d.buffers[name] = &Buffer{Name: name, Visible: true}
	}
	d.current = name
	d.ShowBufferContext()
}

func (d *Display) ShowBufferContext() {
    d.ShowBuffer(d.current, true)
}

// ShowBuffer displays the contents of the specified buffer
// If showHeader is true, it will show the "=== Buffer: name ===" header
func (d *Display) ShowBuffer(name string, showHeader bool) {
    buf := d.buffers[name]
    start := 0
    if len(buf.Lines) > d.lineCount {
        start = len(buf.Lines) - d.lineCount
    }

    if showHeader {
        fmt.Fprintf(d.output, "\n=== Buffer: %s ===\n", name)
    }
    for _, line := range buf.Lines[start:] {
        fmt.Fprintln(d.output, line)
    }
}

func (d *Display) ListBuffers() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	buffers := make([]string, 0, len(d.buffers))
	for name := range d.buffers {
		buffers = append(buffers, name)
	}
	return buffers
}
