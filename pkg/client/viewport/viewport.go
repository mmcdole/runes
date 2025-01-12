package viewport

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmcdole/runes/pkg/client/buffer"
	"strings"
)

type ViewportManager struct {
	viewport viewport.Model
	buffer   *buffer.Buffer
	config   ViewportConfig
	lastLine int  // Track last known line for efficient updates
}

type ViewportConfig struct {
	Width     int
	Height    int
	ScrollOff int // Lines to keep visible when scrolling
}

func NewViewportManager(buffer *buffer.Buffer, config ViewportConfig) *ViewportManager {
	v := viewport.New(config.Width, config.Height)
	v.KeyMap = viewport.KeyMap{} // Disable default keybindings
	return &ViewportManager{
		viewport: v,
		buffer:   buffer,
		config:   config,
		lastLine: 0,
	}
}

func (vm *ViewportManager) Update(msg tea.Msg) (*ViewportManager, tea.Cmd) {
	var cmd tea.Cmd
	vm.viewport, cmd = vm.viewport.Update(msg)
	return vm, cmd
}

func (vm *ViewportManager) SetContent() {
	totalLines := vm.buffer.Length()
	if totalLines == vm.lastLine {
		return // No new content
	}

	// Only get the lines we need
	start := totalLines - vm.viewport.Height
	if start < 0 {
		start = 0
	}
	visibleLines := vm.buffer.GetLines(start, totalLines)
	content := ""
	if len(visibleLines) > 0 {
		content = strings.Join(visibleLines, "\n")
	}
	
	vm.viewport.SetContent(content)
	vm.lastLine = totalLines
	vm.ScrollToBottom()
}

func (vm *ViewportManager) ScrollTo(position int) {
	vm.viewport.SetYOffset(position)
}

func (vm *ViewportManager) ScrollToBottom() {
	vm.viewport.GotoBottom()
}

func (vm *ViewportManager) IsAtBottom() bool {
	return vm.viewport.AtBottom()
}

func (vm *ViewportManager) GetPosition() int {
	return vm.viewport.YOffset
}

func (vm *ViewportManager) View() string {
	return vm.viewport.View()
}

func (vm *ViewportManager) Height() int {
	return vm.viewport.Height
}

func (vm *ViewportManager) Width() int {
	return vm.viewport.Width
}

func (vm *ViewportManager) SetSize(width, height int) {
	wasAtBottom := vm.IsAtBottom()
	vm.viewport.Width = width
	vm.viewport.Height = height
	
	// After resize, update content with new height
	vm.lastLine = 0  // Force content update
	vm.SetContent()
	
	if wasAtBottom {
		vm.ScrollToBottom()
	}
}
