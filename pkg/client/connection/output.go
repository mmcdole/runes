package connection

import (
	"bytes"

	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/client/types"
	"github.com/mmcdole/runes/pkg/protocol/telnet"
)

// TelnetMode represents the current mode of the telnet connection
type TelnetMode struct{}

// OutputProcessor handles the processing of telnet output into lines
type OutputProcessor struct {
	conn   *telnet.TelnetConnection
	buffer []byte
	events *events.EventProcessor
	mode   TelnetMode
}

// NewOutputProcessor creates a new output processor
func NewOutputProcessor(conn *telnet.TelnetConnection, events *events.EventProcessor) *OutputProcessor {
	return &OutputProcessor{
		conn:   conn,
		events: events,
		buffer: make([]byte, 0, 4096),
		mode:   TelnetMode{},
	}
}

// Start begins processing output from the telnet connection
func (p *OutputProcessor) Start() {
	go p.processOutput()
}

// processOutput reads from the telnet connection and processes the output
func (p *OutputProcessor) processOutput() {
	buf := make([]byte, 4096)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			// Connection closed or error
			p.events.Emit(events.Event{
				Type: events.EventDisconnected,
			})
			return
		}

		lines, prompt := p.Process(buf[:n])
		for _, line := range lines {
			p.emitLine(line)
		}
		if prompt != nil {
			p.emitLine(prompt)
		}
	}
}

// Process processes the given data and returns a list of complete lines and a potential prompt
func (p *OutputProcessor) Process(data []byte) ([]*types.Line, *types.Line) {
	// Append new data to existing buffer
	p.buffer = append(p.buffer, data...)

	var lines []*types.Line
	var prompt *types.Line

	// Find complete lines (ending in \n)
	for {
		i := bytes.IndexByte(p.buffer, '\n')
		if i == -1 {
			break
		}

		// Extract line content
		content := string(p.buffer[:i])
		p.buffer = p.buffer[i+1:]

		// Create line
		line := types.NewLine(content)
		line.Complete = true
		lines = append(lines, line)
	}

	// Check remaining buffer for prompt
	if len(p.buffer) > 0 {
		content := string(p.buffer)
		prompt = types.NewPrompt(content)
		p.buffer = p.buffer[:0]
	}

	return lines, prompt
}

// emitLine creates and emits a line event
func (p *OutputProcessor) emitLine(line *types.Line) {
	eventType := events.EventRawOutput
	if line.IsPrompt {
		eventType = events.EventPrompt
	}

	p.events.Emit(events.Event{
		Type: eventType,
		Data: line,
	})
}

// Close cleans up the output processor
func (p *OutputProcessor) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}
