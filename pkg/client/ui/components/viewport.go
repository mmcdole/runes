package components

import (
    "fmt"
    "github.com/mmcdole/runes/pkg/client/buffer"
    "github.com/mmcdole/runes/pkg/client/terminal"
    "github.com/mmcdole/runes/pkg/client/types"
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
    startLine int
    endLine   int
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
    height := v.region.Height

    if v.mode == ModeNormal {
        // In normal mode, always show the latest lines
        if bufLen <= height {
            v.startLine = 0
        } else {
            v.startLine = bufLen - height
        }
        v.endLine = bufLen
    } else {
        // In scrolling mode, keep current window
        if v.endLine > bufLen {
            v.endLine = bufLen
        }
    }

    v.Render()
}

// ScrollUp moves viewport up through history
func (v *Viewport) ScrollUp() {
    if v.mode == ModeNormal {
        v.mode = ModeScrolling
        v.UpdateView()
        return
    }

    height := v.region.Height
    if v.startLine > 0 {
        v.startLine--
        v.endLine--
    }
    if v.endLine-v.startLine < height {
        if v.endLine < v.buffer.Len() {
            v.endLine++
        }
    }
    v.Render()
}

// ScrollDown moves viewport down through history
func (v *Viewport) ScrollDown() {
    if v.mode != ModeScrolling {
        return
    }

    bufLen := v.buffer.Len()
    if v.endLine < bufLen {
        v.startLine++
        v.endLine++
    }
    if v.endLine == bufLen {
        v.ScrollToBottom()
    } else {
        v.Render()
    }
}

// ScrollToTop scrolls to the start of history
func (v *Viewport) ScrollToTop() {
    if v.mode == ModeNormal {
        v.mode = ModeScrolling
    }
    v.startLine = 0
    v.endLine = v.region.Height
    if v.endLine > v.buffer.Len() {
        v.endLine = v.buffer.Len()
    }
    v.Render()
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
func (v *Viewport) Render() {
    // Get lines including prompt if present
    lines := v.buffer.GetLines(v.startLine, v.endLine)
    for i, line := range lines {
        v.renderLine(i, line)
    }
}

func (v *Viewport) renderLine(index int, line types.Line) {
    // Skip gagged lines
    if line.Gag {
        return
    }

    // Clear line and move cursor
    output := fmt.Sprintf("\033[%d;%dH\033[K", v.region.Row+index+1, v.region.Col+1)
    
    // Write display content (already ANSI processed)
    output += line.Display
    
    // Send all operations in a single write
    v.term.Write([]byte(output))
}
