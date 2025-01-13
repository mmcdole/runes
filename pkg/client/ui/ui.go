package ui

import (
	"fmt"
	"sync"

	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/screen"
)

// UI represents the terminal user interface
type UI struct {
	screen     *screen.Screen
	client     Client
	buffer     *buffer.BufferManager
	done       chan struct{}
	inputMutex sync.Mutex
	inputBuf   string
	inputPos   int
	statusText string
}

// Client interface for UI callbacks
type Client interface {
	HandleInput(input string)
	HandleQuit()
}

const (
	statusHeight = 1
	inputHeight  = 1
)

// New creates a new UI instance
func New(client Client, bufferMgr *buffer.BufferManager, opts ...Option) (*UI, error) {
	scr, err := screen.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	ui := &UI{
		screen:  scr,
		client:  client,
		buffer:  bufferMgr,
		done:    make(chan struct{}),
	}

	// Register draw callback for buffer updates
	bufferMgr.SetUpdateCallback(ui.draw)

	// Start event handling
	go ui.handleEvents()

	for _, opt := range opts {
		opt(ui)
	}

	return ui, nil
}

// Run starts the UI
func (ui *UI) Run() error {
	// Initialize screen regions
	_, height := ui.screen.Size()
	ui.screen.SetScrollRegion(statusHeight, height-inputHeight-1)
	
	// Hide cursor during initial draw
	ui.screen.ShowCursor(false)
	
	// Draw initial screen
	ui.drawStatus()
	ui.drawBuffer()
	ui.drawInput()
	
	// Show cursor at input line
	ui.screen.ShowCursor(true)
	return nil
}

// SetStatus sets the status text
func (ui *UI) SetStatus(text string) {
	ui.statusText = text
	ui.drawStatus()
}

// PrintOutput prints a line to the output area
func (ui *UI) PrintOutput(line string) {
	ui.buffer.AddLine(line)
	ui.drawBuffer()
}

// drawStatus draws the status bar
func (ui *UI) drawStatus() {
	width, _ := ui.screen.Size()
	ui.screen.WriteAt(fmt.Sprintf("\x1b[44m%-*s\x1b[0m", width, ui.statusText), 0, 0)
}

// drawBuffer draws the buffer content
func (ui *UI) drawBuffer() {
	_, height := ui.screen.Size()
	contentHeight := height - statusHeight - inputHeight

	// Get visible lines
	buf := ui.buffer.GetCurrentBuffer()
	lines := buf.GetLines(buf.Length()-contentHeight, buf.Length())

	// Clear content area
	ui.screen.ClearRegion(statusHeight, height-inputHeight-1)

	// Draw visible lines
	for i, line := range lines {
		ui.screen.WriteAt(line, 0, statusHeight+i)
	}
}

// drawInput draws the input line
func (ui *UI) drawInput() {
	_, height := ui.screen.Size()
	inputLine := height - 1

	// Save cursor position
	ui.screen.SaveCursor()

	// Clear and draw input line
	ui.screen.WriteAt("\x1b[40m> "+ui.inputBuf+"\x1b[0m", 0, inputLine)
	ui.screen.SetCursor(ui.inputPos+2, inputLine)

	// Restore cursor position
	ui.screen.RestoreCursor()
}

// draw updates the entire UI
func (ui *UI) draw() {
	ui.screen.ShowCursor(false)
	ui.drawStatus()
	ui.drawBuffer()
	ui.drawInput()
	ui.screen.ShowCursor(true)
}

// handleEvents processes UI events
func (ui *UI) handleEvents() {
	for {
		select {
		case <-ui.done:
			return
		default:
			event := ui.screen.PollEvent()
			switch event.Type {
			case screen.KeyTypeResize:
				_, height := ui.screen.Size()
				ui.screen.SetScrollRegion(statusHeight, height-inputHeight-1)
				ui.draw()
			case screen.KeyTypeSpecial:
				switch event.Special {
				case screen.KeyCtrlC:
					ui.client.HandleQuit()
					return
				case screen.KeyEnter:
					if len(ui.inputBuf) > 0 {
						ui.client.HandleInput(ui.inputBuf)
						ui.inputBuf = ""
						ui.inputPos = 0
						ui.drawInput()
					}
				case screen.KeyBackspace:
					if ui.inputPos > 0 {
						ui.inputBuf = ui.inputBuf[:ui.inputPos-1] + ui.inputBuf[ui.inputPos:]
						ui.inputPos--
						ui.drawInput()
					}
				case screen.KeyDelete:
					if ui.inputPos < len(ui.inputBuf) {
						ui.inputBuf = ui.inputBuf[:ui.inputPos] + ui.inputBuf[ui.inputPos+1:]
						ui.drawInput()
					}
				case screen.KeyLeft:
					if ui.inputPos > 0 {
						ui.inputPos--
						ui.drawInput()
					}
				case screen.KeyRight:
					if ui.inputPos < len(ui.inputBuf) {
						ui.inputPos++
						ui.drawInput()
					}
				case screen.KeyHome:
					ui.inputPos = 0
					ui.drawInput()
				case screen.KeyEnd:
					ui.inputPos = len(ui.inputBuf)
					ui.drawInput()
				}
			case screen.KeyTypeRune:
				if ui.inputPos == len(ui.inputBuf) {
					ui.inputBuf += string(event.Rune)
				} else {
					ui.inputBuf = ui.inputBuf[:ui.inputPos] + string(event.Rune) + ui.inputBuf[ui.inputPos:]
				}
				ui.inputPos++
				ui.drawInput()
			}
		}
	}
}

// Close cleans up the UI
func (ui *UI) Close() {
	close(ui.done)
	ui.screen.Close()
}

type Option func(*UI)
