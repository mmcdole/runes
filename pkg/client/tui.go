package client

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	viewport   viewport.Model
	input      textinput.Model
	bufferMgr  *BufferManager
	client     *Client
	ready      bool
	windowSize tea.WindowSizeMsg
}

func NewModel(client *Client, bufferMgr *BufferManager) *model {
	input := textinput.New()
	input.Focus()
	input.Prompt = "> "
	input.Width = 80

	return &model{
		input:     input,
		bufferMgr: bufferMgr,
		client:    client,
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Send the command to the client
			input := m.input.Value()
			m.client.HandleInput(input)
			m.input.Reset()
			return m, textinput.Blink

		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-2)
			m.viewport.SetContent(m.bufferContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
		}
		m.windowSize = msg
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if !m.ready {
		return "\nInitializing..."
	}

	return fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.input.View(),
	)
}

func (m *model) bufferContent() string {
	buffer := m.bufferMgr.GetCurrentBuffer()
	content := ""
	for _, line := range buffer.Lines {
		content += line + "\n"
	}
	return content
}

func (m *model) UpdateContent() {
	if m.ready {
		m.viewport.SetContent(m.bufferContent())
		m.viewport.GotoBottom()
	}
}
