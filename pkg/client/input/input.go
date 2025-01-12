package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputManager struct {
	input    textinput.Model
	history  []string
	position int
	config   InputConfig
}

type InputConfig struct {
	Prompt        string
	MaxHistory    int
	InitialWidth  int
}

func NewInputManager(config InputConfig) *InputManager {
	input := textinput.New()
	input.Prompt = config.Prompt
	input.Width = config.InitialWidth
	input.Focus()

	return &InputManager{
		input:    input,
		history:  make([]string, 0, config.MaxHistory),
		position: -1,
		config:   config,
	}
}

func (im *InputManager) HandleInput(msg tea.Msg) (string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if im.input.Value() == "" {
				return "", nil
			}
			input := im.input.Value()
			im.AddToHistory(input)
			im.input.Reset()
			return input, nil
		case tea.KeyUp:
			im.Previous()
			return "", nil
		case tea.KeyDown:
			im.Next()
			return "", nil
		}
	}

	var cmd tea.Cmd
	im.input, cmd = im.input.Update(msg)
	return "", cmd
}

func (im *InputManager) AddToHistory(input string) {
	if len(im.history) >= im.config.MaxHistory {
		// Remove oldest entry
		im.history = append(im.history[1:], input)
	} else {
		im.history = append(im.history, input)
	}
	im.position = len(im.history)
}

func (im *InputManager) Previous() {
	if len(im.history) == 0 || im.position <= 0 {
		return
	}
	
	if im.position == len(im.history) {
		// Save current input if we're starting to navigate history
		im.history = append(im.history, im.input.Value())
	}
	
	im.position--
	im.input.SetValue(im.history[im.position])
}

func (im *InputManager) Next() {
	if im.position >= len(im.history)-1 {
		return
	}
	im.position++
	im.input.SetValue(im.history[im.position])
}

func (im *InputManager) View() string {
	return im.input.View()
}

func (im *InputManager) SetWidth(width int) {
	im.input.Width = width
}

func (im *InputManager) Focus() {
	im.input.Focus()
}

func (im *InputManager) Blur() {
	im.input.Blur()
}
