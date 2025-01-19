package components

import (
	"fmt"
	"strings"
	"github.com/mmcdole/runes/pkg/client/terminal"
	"github.com/mmcdole/runes/pkg/client/ui/layout"
)

// StatusBar represents the status line at the top of the screen
type StatusBar struct {
	term    *terminal.Terminal
	text    string
	region  layout.Region
}

// NewStatusBar creates a new status bar
func NewStatusBar(term *terminal.Terminal) *StatusBar {
	return &StatusBar{
		term: term,
		region: layout.Region{
			Row:    1,    // ANSI escape codes use 1-based indexing
			Col:    1,    // ANSI escape codes use 1-based indexing
			Height: 1,
		},
	}
}

// SetText updates the status bar text
func (s *StatusBar) SetText(text string) {
	s.text = text
	s.Render(s.region)
}

// Render draws the status bar
func (s *StatusBar) Render(region layout.Region) {
	s.region = region
	
	// Ensure we have valid dimensions
	if region.Width < 1 {
		region.Width = 1
	}
	
	// Disable origin mode to use absolute positioning
	s.term.Write([]byte("\033[?6l"))

	// Move to status line position
	s.term.Write([]byte(fmt.Sprintf("\033[%d;%dH", region.Row+1, region.Col+1)))
	
	// Set background color (dark gray) and white text
	s.term.Write([]byte("\033[48;5;237m\033[37m"))

	// Clear the line
	s.term.Write([]byte("\033[K"))

	// Add some padding around the text
	paddedText := fmt.Sprintf(" %s ", s.text)
	
	// Truncate text if it's too long for the region
	if len(paddedText) > region.Width {
		paddedText = paddedText[:region.Width]
	}

	// Write text and pad the rest of the line with spaces
	s.term.Write([]byte(paddedText))
	if padding := region.Width - len(paddedText); padding > 0 {
		s.term.Write([]byte(strings.Repeat(" ", padding)))
	}

	// Reset colors
	s.term.Write([]byte("\033[0m"))
}
