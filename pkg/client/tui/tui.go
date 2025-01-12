package tui

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/rivo/tview"
)

// TUI represents the terminal user interface
type TUI struct {
	app        *tview.Application
	textView   *tview.TextView
	inputField *tview.InputField
	buffer     *buffer.Buffer
	client     Client

	// Mutex for synchronizing buffer access
	bufferMutex sync.Mutex

	// Channel for serializing output updates
	outputChan chan string
}

// Config holds the TUI configuration
type Config struct {
	BufferSize int
}

// Client interface for handling input and quit events
type Client interface {
	HandleInput(input string)
	HandleQuit()
}

// New creates a new TUI instance
func New(config Config, client Client) *TUI {
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		SetWrap(true)

	inputField := tview.NewInputField().
		SetLabel("> ")

	t := &TUI{
		app:        app,
		textView:   textView,
		inputField: inputField,
		client:     client,
		buffer:     buffer.NewBuffer(buffer.BufferConfig{MaxLines: config.BufferSize}),
		outputChan: make(chan string, 1000), // Buffer size to prevent blocking
	}

	// Set up input handling
	inputField.SetFinishedFunc(func(key tcell.Key) {
		text := inputField.GetText()
		if text != "" {
			line := fmt.Sprintf("> %s\n", text)
			t.bufferMutex.Lock()
			t.buffer.Append(line)
			t.bufferMutex.Unlock()

			// UI updates are safe here because we're in the main thread
			textView.Write([]byte(line))
			textView.ScrollToEnd()
			inputField.SetText("")
			t.client.HandleInput(text)
		}
	})

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(inputField, 1, 0, true)

	// Set up global key handlers
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			t.client.HandleQuit()
			app.Stop()
			return nil
		case tcell.KeyPgUp:
			row, _ := textView.GetScrollOffset()
			textView.ScrollTo(row-1, 0)
			app.Draw()
			return nil
		case tcell.KeyPgDn:
			row, _ := textView.GetScrollOffset()
			textView.ScrollTo(row+1, 0)
			app.Draw()
			return nil
		}
		return event
	})

	app.SetRoot(flex, true).
		EnableMouse(true).
		SetFocus(inputField)

	// Start output processor
	go t.processOutput()

	return t
}

// Convert ANSI escape sequences to tview color tags
func convertANSIToTview(text string) string {
	// ANSI escape sequence pattern
	pattern := regexp.MustCompile(`\x1b\[([0-9;]*)m`)

	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		code := pattern.FindStringSubmatch(match)[1]
		switch code {
		case "0":
			return "[-:-:-]" // Reset
		case "30":
			return "[black]"
		case "31":
			return "[red]"
		case "32":
			return "[green]"
		case "33":
			return "[yellow]"
		case "34":
			return "[blue]"
		case "35":
			return "[purple]"
		case "36":
			return "[cyan]"
		case "37":
			return "[white]"
		default:
			return "" // Remove unknown codes
		}
	})
}

// processOutput handles output events serially
func (t *TUI) processOutput() {
	for line := range t.outputChan {
		t.bufferMutex.Lock()
		t.buffer.Append(line)
		t.bufferMutex.Unlock()

		t.app.QueueUpdateDraw(func() {
			// Convert ANSI colors to tview format
			displayLine := convertANSIToTview(line)
			t.textView.Write([]byte(displayLine))
			t.textView.ScrollToEnd()
		})
	}
}

// Run starts the TUI
func (t *TUI) Run() error {
	return t.app.Run()
}

// Stop stops the TUI
func (t *TUI) Stop() {
	close(t.outputChan)
	t.app.Stop()
}

// AddLine adds a line to the output
func (t *TUI) AddLine(line string) {
	// Ensure newline
	if line[len(line)-1] != '\n' {
		line = line + "\n"
	}

	// Send to output processor
	t.outputChan <- line
}

// Clear clears all output
func (t *TUI) Clear() {
	t.bufferMutex.Lock()
	t.buffer.Clear()
	t.bufferMutex.Unlock()

	t.app.QueueUpdateDraw(func() {
		t.textView.Clear()
	})
}
