package mud

import (
	"fmt"
	"io"
	"sync"
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
		current:   "main",
		lineCount: 50,
	}
	d.buffers["main"] = &Buffer{Name: "main", Visible: true}
	d.buffers["system"] = &Buffer{Name: "system", Visible: true}
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
	buf := d.buffers[d.current]
	start := 0
	if len(buf.Lines) > d.lineCount {
		start = len(buf.Lines) - d.lineCount
	}

	fmt.Fprintf(d.output, "\n=== Buffer: %s ===\n", d.current)
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
