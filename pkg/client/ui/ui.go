package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/screen"
)

// UI represents the terminal user interface
type UI struct {
	screen     *screen.Screen
	client     Client
	buffer     *buffer.BufferManager
	autoScroll bool
	statusText string
	scrollPos  int

	// Input handling
	inputBuf   string
	inputPos   int
	inputMutex sync.Mutex

	// Event channels
	done chan struct{}
}

// Client interface for handling input and quit events
type Client interface {
	HandleInput(input string)
	HandleQuit()
}

// Config holds the UI configuration
type Config struct {
	BufferSize int
}

// New creates a new UI instance
func New(config Config, client Client, bufferMgr *buffer.BufferManager) (*UI, error) {
	scr, err := screen.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	ui := &UI{
		screen:     scr,
		client:     client,
		buffer:     bufferMgr,
		done:       make(chan struct{}),
		autoScroll: true,
	}

	// Register draw callback for buffer updates
	bufferMgr.SetUpdateCallback(ui.draw)

	// Start event handling
	go ui.handleEvents()

	return ui, nil
}

// handleEvents processes terminal events
func (ui *UI) handleEvents() {
	for {
		select {
		case <-ui.done:
			return
		default:
			ev := ui.screen.PollEvent()
			ui.handleKeyEvent(ev)
		}
	}
}

// handleKeyEvent processes keyboard input
func (ui *UI) handleKeyEvent(ev screen.Key) {
	ui.inputMutex.Lock()
	defer ui.inputMutex.Unlock()

	switch ev.Type {
	case screen.KeyTypeSpecial:
		switch ev.Special {
		case screen.KeyEnter:
			if ui.inputBuf != "" {
				input := ui.inputBuf
				ui.inputBuf = ""
				ui.inputPos = 0
				ui.draw()
				ui.client.HandleInput(input)
			}
		case screen.KeyBackspace:
			if ui.inputPos > 0 {
				ui.inputBuf = ui.inputBuf[:ui.inputPos-1] + ui.inputBuf[ui.inputPos:]
				ui.inputPos--
			}
		case screen.KeyDelete:
			if ui.inputPos < len(ui.inputBuf) {
				ui.inputBuf = ui.inputBuf[:ui.inputPos] + ui.inputBuf[ui.inputPos+1:]
			}
		case screen.KeyLeft:
			if ui.inputPos > 0 {
				ui.inputPos--
			}
		case screen.KeyRight:
			if ui.inputPos < len(ui.inputBuf) {
				ui.inputPos++
			}
		case screen.KeyHome:
			ui.inputPos = 0
		case screen.KeyEnd:
			ui.inputPos = len(ui.inputBuf)
		case screen.KeyPgUp:
			ui.ScrollUp(10)
		case screen.KeyPgDn:
			ui.ScrollDown(10)
		case screen.KeyCtrlC:
			ui.client.HandleQuit()
		}
	case screen.KeyTypeRune:
		// Insert character at cursor position
		if ui.inputPos == len(ui.inputBuf) {
			ui.inputBuf += string(ev.Rune)
		} else {
			ui.inputBuf = ui.inputBuf[:ui.inputPos] + string(ev.Rune) + ui.inputBuf[ui.inputPos:]
		}
		ui.inputPos++
	}

	ui.draw()
}

// draw renders the entire UI
func (ui *UI) draw() {
	_, height := ui.screen.Size()
	
	// Clear entire screen first
	ui.screen.Clear()
	
	// Draw status bar with background color
	ui.screen.Write(fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m\r\n", 37, 44, ui.statusText))

	// Get visible lines
	buf := ui.buffer.GetCurrentBuffer()
	maxScroll := buf.Length() - (height - 2) // -2 for status and input lines
	if maxScroll < 0 {
		maxScroll = 0
	}

	// If auto-scroll is enabled, move to the bottom
	if ui.autoScroll {
		ui.scrollPos = maxScroll
	}

	// Get the lines to display
	lines := buf.GetLines(ui.scrollPos, ui.scrollPos+(height-2))

	// Write each line with newline
	for _, line := range lines {
		ui.screen.Write(line + "\r\n")
	}

	// Move to input line and clear it
	ui.screen.SetCursor(0, height-1)
	ui.screen.Write("\x1b[2K") // Clear entire line
		
	// Draw input line with prompt
	ui.screen.Write(fmt.Sprintf("\x1b[%d;%dm> %s\x1b[0m", 37, 40, ui.inputBuf))
	ui.screen.SetCursor(ui.inputPos+2, height-1)
}

// AddLine adds a line to the output
func (ui *UI) AddLine(line string) {
	ui.inputMutex.Lock()
	defer ui.inputMutex.Unlock()

	// Split the line by newlines and add each part
	parts := strings.Split(line, "\n")
	for _, l := range parts {
		if l != "" {
			ui.buffer.AddLine(l)
		}
	}

	ui.draw()
}

// SetStatus sets the status bar text
func (ui *UI) SetStatus(text string) {
	ui.inputMutex.Lock()
	defer ui.inputMutex.Unlock()
	ui.statusText = text
	ui.draw()
}

// Run starts the UI
func (ui *UI) Run() error {
	// Make sure cursor is hidden
	ui.screen.ShowCursor(false)
	
	// Draw initial screen
	ui.draw()
	
	// Show cursor at input line
	ui.screen.ShowCursor(true)
	return nil
}

// Stop stops the UI
func (ui *UI) Stop() {
	close(ui.done)
	ui.screen.Close()
}

// Clear clears the viewport
func (ui *UI) Clear() {
	ui.screen.Clear()
	ui.draw()
}

// ScrollUp scrolls the content up
func (ui *UI) ScrollUp(lines int) {
	ui.inputMutex.Lock()
	defer ui.inputMutex.Unlock()
	
	ui.scrollPos -= lines
	if ui.scrollPos < 0 {
		ui.scrollPos = 0
	}
	ui.autoScroll = false
	ui.draw()
}

// ScrollDown scrolls the content down
func (ui *UI) ScrollDown(lines int) {
	ui.inputMutex.Lock()
	defer ui.inputMutex.Unlock()

	_, height := ui.screen.Size()
	buf := ui.buffer.GetCurrentBuffer()
	maxScroll := buf.Length() - (height - 2)
	if maxScroll < 0 {
		maxScroll = 0
	}

	ui.scrollPos += lines
	if ui.scrollPos > maxScroll {
		ui.scrollPos = maxScroll
		ui.autoScroll = true
	}
	ui.draw()
}

// HandleResize handles terminal resize events
func (ui *UI) HandleResize() {
	ui.inputMutex.Lock()
	defer ui.inputMutex.Unlock()
	ui.draw()
}
