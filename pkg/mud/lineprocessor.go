package mud

import (
	"bytes"
)

// LineProcessor processes incoming data into lines
type LineProcessor struct{}

// NewLineProcessor returns a new LineProcessor
func NewLineProcessor() *LineProcessor {
	return &LineProcessor{}
}

// Write processes incoming data and returns lines
func (p *LineProcessor) Write(data []byte) []string {
	var lines []string
	buf := data

	for {
		idx := bytes.IndexByte(buf, '\n')
		if idx == -1 {
			// No more newlines, send remaining data as a line
			if len(buf) > 0 {
				lines = append(lines, string(buf))
			}
			return lines
		}

		// Extract the line (including any \r)
		line := buf[:idx]
		if idx > 0 && line[idx-1] == '\r' {
			line = line[:idx-1]
		}
		lines = append(lines, string(line))

		// Move forward
		buf = buf[idx+1:]
	}
}
