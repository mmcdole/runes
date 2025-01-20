package connection

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/protocol/telnet"
)

type mockTelnetConnection struct {
	*telnet.TelnetConnection
	readData []byte
	readErr  error
}

func newMockConnection() *mockTelnetConnection {
	return &mockTelnetConnection{
		TelnetConnection: &telnet.TelnetConnection{},
	}
}

func (m *mockTelnetConnection) Read(p []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	n = copy(p, m.readData)
	return n, nil
}

func (m *mockTelnetConnection) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockTelnetConnection) Close() error {
	return nil
}

func collectLines(t *testing.T, eventProcessor *events.EventProcessor, timeout time.Duration, expectedLines int) []*Line {
	t.Helper()
	lines := make([]*Line, 0)
	done := make(chan struct{})
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	eventProcessor.Subscribe(events.EventRawOutput, func(e events.Event) {
		line := e.Data.(*Line)
		fmt.Printf("Got line: %+v\n", line)
		lines = append(lines, line)
		// For tests that expect only one line, we can stop waiting
		if len(lines) == expectedLines {
			done <- struct{}{}
		}
	})

	// Wait for either timeout or first line
	select {
	case <-done:
	case <-timer.C:
	}

	return lines
}

func TestOutputProcessor_ProcessBuffer(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantLine string
		complete bool
	}{
		{
			name:     "complete line with newline",
			input:    []byte("test\n"),
			wantLine: "test",
			complete: true,
		},
		{
			name:     "incomplete line",
			input:    []byte("test"),
			wantLine: "test",
			complete: false,
		},
		{
			name:     "line with CRLF",
			input:    []byte("test\r\n"),
			wantLine: "test",
			complete: true,
		},
		{
			name:     "multiple complete lines",
			input:    []byte("line1\nline2\n"),
			wantLine: "line2",
			complete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newMockConnection()
			eventProcessor := events.New()
			proc := NewOutputProcessor(conn.TelnetConnection, eventProcessor)

			fmt.Printf("Processing input: %q\n", tt.input)
			proc.Write(tt.input)

			// Collect lines
			gotLines := collectLines(t, eventProcessor, 100*time.Millisecond, 1)

			if len(gotLines) == 0 {
				t.Fatal("no lines received")
			}

			lastLine := gotLines[len(gotLines)-1]
			if lastLine.Content != tt.wantLine {
				t.Errorf("got line content %q, want %q", lastLine.Content, tt.wantLine)
			}
			if lastLine.Complete != tt.complete {
				t.Errorf("got complete %v, want %v", lastLine.Complete, tt.complete)
			}
		})
	}
}
