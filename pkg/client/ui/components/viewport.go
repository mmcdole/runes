package components

import (
	"fmt"
	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/terminal"
	"github.com/mmcdole/runes/pkg/client/ui/layout"
)

// ViewportMode represents the current mode of the viewport
type ViewportMode int

const (
	ModeNormal ViewportMode = iota
	ModeScrolling
)

// String returns the string representation of the mode
func (m ViewportMode) String() string {
	switch m {
	case ModeNormal:
		return "Live"
	case ModeScrolling:
		return "Scrolling"
	default:
		return "Unknown"
	}
}

// Viewport represents the main content area
type Viewport struct {
	term      *terminal.Terminal
	buffer    *buffer.Buffer
	mode      ViewportMode
	region    layout.Region
	visibleRange struct {
		start int
		count int
	}
}

// NewViewport creates a new viewport
func NewViewport(term *terminal.Terminal, buf *buffer.Buffer) *Viewport {
	return &Viewport{
		term:   term,
		buffer: buf,
		mode:   ModeNormal,
	}
}

// UpdateView recalculates what should be visible and renders
func (v *Viewport) UpdateView() {
	bufLen := v.buffer.Len()
	
	if v.mode == ModeNormal {
		// In normal mode, show the most recent lines
		if bufLen > v.region.Height {
			v.visibleRange.start = bufLen - v.region.Height
		} else {
			v.visibleRange.start = 0
		}
		v.visibleRange.count = v.region.Height
	} else {
		// In scroll mode, maintain current position
		v.visibleRange.count = v.region.Height
	}
	
	v.Render(v.region)
}

// ScrollUp moves viewport up through history
func (v *Viewport) ScrollUp() {
	bufLen := v.buffer.Len()
	
	if v.mode == ModeNormal {
		// When first entering scroll mode, scroll back one line
		v.mode = ModeScrolling
		v.visibleRange.start = bufLen - v.region.Height - 1
		if v.visibleRange.start < 0 {
			v.visibleRange.start = 0
		}
	} else {
		if v.visibleRange.start > 0 {
			v.visibleRange.start--
		}
	}
	v.Render(v.region)
}

// ScrollDown moves viewport down through history
func (v *Viewport) ScrollDown() {
	if v.mode != ModeScrolling {
		return
	}
	
	maxStart := v.buffer.Len() - v.region.Height
	if v.visibleRange.start < maxStart {
		v.visibleRange.start++
		v.Render(v.region)
	} else {
		v.ScrollToBottom()
	}
}

// ScrollToTop scrolls to the start of history
func (v *Viewport) ScrollToTop() {
	if v.buffer.Len() > v.region.Height {
		v.mode = ModeScrolling
	}
	v.visibleRange.start = 0
	v.Render(v.region)
}

// ScrollToBottom returns to live mode at the bottom
func (v *Viewport) ScrollToBottom() {
	v.mode = ModeNormal
	v.UpdateView()
}

// GetMode returns the current viewport mode
func (v *Viewport) GetMode() ViewportMode {
	return v.mode
}

// Render draws the viewport content
func (v *Viewport) Render(region layout.Region) {
	v.region = region
	
	// Clear viewport area
	for i := 0; i < region.Height; i++ {
		v.term.Write([]byte(fmt.Sprintf("\033[%d;%dH\033[K", region.Row+i+1, region.Col+1)))
	}
	
	// Get visible lines
	lines := v.buffer.GetLines(v.visibleRange.start, v.visibleRange.count)
	
	// Write visible content
	for i, line := range lines {
		v.term.Write([]byte(fmt.Sprintf("\033[%d;%dH%s", region.Row+i+1, region.Col+1, line.Content)))
	}
}
