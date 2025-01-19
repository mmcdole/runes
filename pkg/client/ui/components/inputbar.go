package components

import (
	"fmt"
	"github.com/mmcdole/runes/pkg/client/history"
	"github.com/mmcdole/runes/pkg/client/terminal"
	"github.com/mmcdole/runes/pkg/client/ui/layout"
)

const (
	defaultPrompt = "> "
)

// InputBar handles user input with history browsing capabilities
type InputBar struct {
	term        *terminal.Terminal
	content     string
	position    int
	savedInput  string
	history     *history.History
	historyPos  int  // -1 means not in history mode
	region      layout.Region
	prompt      string
}

// NewInputBar creates a new input bar component
func NewInputBar(term *terminal.Terminal, history *history.History) *InputBar {
	return &InputBar{
		term:       term,
		history:    history,
		historyPos: -1,
		prompt:     defaultPrompt,
	}
}

// SetPrompt sets the prompt string
func (b *InputBar) SetPrompt(prompt string) {
	b.prompt = prompt
	b.Render(b.region)
}

// HandleInput processes input and returns true if handled
func (b *InputBar) HandleInput(input []byte) bool {
	// Handle Enter key
	if terminal.IsEnter(input) {
		// Return true to indicate we want to submit
		return true
	}

	// Handle history navigation
	if terminal.IsUpArrow(input) {
		return b.browseUp()
	}
	if terminal.IsDownArrow(input) {
		return b.browseDown()
	}

	if terminal.IsBackspace(input) {
		if b.position > 0 {
			b.content = b.content[:b.position-1] + b.content[b.position:]
			b.position--
			b.Render(b.region)
		}
		return true
	}

	if terminal.IsDelete(input) {
		if b.position < len(b.content) {
			b.content = b.content[:b.position] + b.content[b.position+1:]
			b.Render(b.region)
		}
		return true
	}

	if terminal.IsLeftArrow(input) {
		if b.position > 0 {
			b.position--
			b.Render(b.region)
		}
		return true
	}

	if terminal.IsRightArrow(input) {
		if b.position < len(b.content) {
			b.position++
			b.Render(b.region)
		}
		return true
	}

	if terminal.IsHome(input) {
		b.position = 0
		b.Render(b.region)
		return true
	}

	if terminal.IsEnd(input) {
		b.position = len(b.content)
		b.Render(b.region)
		return true
	}

	// Regular text input
	if len(input) == 1 && input[0] >= 32 && input[0] <= 126 {
		if b.position == len(b.content) {
			b.content += string(input[0])
		} else {
			b.content = b.content[:b.position] + string(input[0]) + b.content[b.position:]
		}
		b.position++
		b.Render(b.region)
		return true
	}

	return false
}

// browseUp moves to the previous command in history
func (b *InputBar) browseUp() bool {
	if b.historyPos == -1 {
		// Starting history browsing
		b.historyPos = b.history.Length()
		b.savedInput = b.content
	}

	if b.historyPos > 0 {
		b.historyPos--
		if cmd, ok := b.history.Get(b.historyPos); ok {
			b.SetContent(cmd)
			return true
		}
	}
	return false
}

// browseDown moves to the next command in history
func (b *InputBar) browseDown() bool {
	if b.historyPos == -1 {
		return false
	}

	b.historyPos++
	if b.historyPos >= b.history.Length() {
		// Reached the end, restore original input
		b.historyPos = -1
		b.SetContent(b.savedInput)
		b.savedInput = ""
		return true
	}

	if cmd, ok := b.history.Get(b.historyPos); ok {
		b.SetContent(cmd)
		return true
	}
	return false
}

// GetContent returns the current input content
func (b *InputBar) GetContent() string {
	return b.content
}

// SetContent updates the input content and moves cursor to end
func (b *InputBar) SetContent(content string) {
	b.content = content
	b.position = len(content)
	b.Render(b.region)
}

// Clear clears the input content
func (b *InputBar) Clear() {
	b.content = ""
	b.position = 0
	b.savedInput = ""
	b.historyPos = -1
	b.Render(b.region)
}

// ExitHistory exits history browsing mode
func (b *InputBar) ExitHistory() {
	b.historyPos = -1
	b.savedInput = ""
}

// Render implements layout.Component interface
func (b *InputBar) Render(region layout.Region) {
	b.region = region
	
	// Move cursor to input bar position (add 1 for 1-based cursor)
	b.term.MoveCursor(region.Row+1, region.Col+1)
	
	// Clear the line
	b.term.ClearLine()
	
	// Write prompt and content
	fmt.Fprintf(b.term, "%s%s", b.prompt, b.content)
	
	// Position cursor (add 1 for 1-based cursor)
	b.term.MoveCursor(region.Row+1, region.Col+len(b.prompt)+b.position+1)
}
