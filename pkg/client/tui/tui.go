package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/input"
	"github.com/mmcdole/runes/pkg/client/viewport"
)

type TUI struct {
	viewport    *viewport.ViewportManager
	input       *input.InputManager
	buffer      *buffer.Buffer
	ready       bool
	config      Config
	client      Client
}

type Config struct {
	BufferConfig   buffer.BufferConfig
	ViewportConfig viewport.ViewportConfig
	InputConfig    input.InputConfig
}

// Client interface represents the external client that handles actual MUD communication
type Client interface {
	HandleInput(string)
	HandleQuit()
}

func New(config Config, client Client) *TUI {
	buf := buffer.NewBuffer(config.BufferConfig)
	vp := viewport.NewViewportManager(buf, config.ViewportConfig)
	inp := input.NewInputManager(config.InputConfig)

	return &TUI{
		viewport: vp,
		input:    inp,
		buffer:   buf,
		config:   config,
		client:   client,
	}
}

func (t *TUI) Init() tea.Cmd {
	return textinput.Blink
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !t.ready {
			// First time initialization
			t.viewport.SetSize(msg.Width, msg.Height-1) // Reserve bottom line for input
			t.input.SetWidth(msg.Width)
			t.ready = true
		} else {
			// Subsequent resize events
			t.viewport.SetSize(msg.Width, msg.Height-1)
			t.input.SetWidth(msg.Width)
		}
		t.viewport.SetContent()

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			t.client.HandleQuit()
			return t, tea.Quit
		}
		
		// Handle input first
		if input, cmd := t.input.HandleInput(msg); input != "" {
			// Send to client
			t.client.HandleInput(input)
			// Echo to our buffer
			t.buffer.Append(fmt.Sprintf("> %s", input))
			t.viewport.SetContent()
			cmds = append(cmds, cmd)
		} else if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update viewport
	if _, cmd = t.viewport.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func (t *TUI) View() string {
	if !t.ready {
		return "\nInitializing..."
	}

	return fmt.Sprintf("%s\n%s",
		t.viewport.View(),
		t.input.View())
}

// AddLine adds a new line to the buffer and updates the viewport
func (t *TUI) AddLine(line string) {
	t.buffer.Append(line)
	if t.ready {
		t.viewport.SetContent()
	}
}

// Clear clears the buffer and updates the viewport
func (t *TUI) Clear() {
	t.buffer.Clear()
	if t.ready {
		t.viewport.SetContent()
	}
}
