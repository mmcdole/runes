package screen

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode/utf8"

	"golang.org/x/term"
)

type KeyType int

const (
	KeyTypeRune KeyType = iota
	KeyTypeSpecial
	KeyTypeResize
)

// Key represents a keyboard input
type Key struct {
	Type    KeyType
	Rune    rune
	Special SpecialKey
}

type SpecialKey int

const (
	KeyNone SpecialKey = iota
	KeyEnter
	KeyBackspace
	KeyDelete
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDn
	KeyCtrlC
)

// Screen handles raw terminal I/O
type Screen struct {
	in           *os.File    // Terminal input
	out          *os.File    // Terminal output
	oldState     *term.State // Original terminal state
	width        int
	height       int
	mu           sync.Mutex
	events       chan Key
	done         chan struct{}
	sigwinch     chan os.Signal // Window size change events
	scrollRegion struct {
		top    int
		bottom int
	}
}

// New creates a new screen
func New() (*Screen, error) {
	// Open terminal for input
	in, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open terminal: %w", err)
	}

	// Save terminal state
	oldState, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		in.Close()
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}

	// Get terminal size
	width, height, err := term.GetSize(int(in.Fd()))
	if err != nil {
		term.Restore(int(in.Fd()), oldState)
		in.Close()
		return nil, fmt.Errorf("failed to get terminal size: %w", err)
	}

	s := &Screen{
		in:       in,
		out:      os.Stdout,
		oldState: oldState,
		width:    width,
		height:   height,
		events:   make(chan Key),
		done:     make(chan struct{}),
		sigwinch: make(chan os.Signal, 1),
	}

	// Set up SIGWINCH handler
	signal.Notify(s.sigwinch, syscall.SIGWINCH)

	// Switch to alternate screen
	fmt.Fprint(s.out, "\x1b[?1049h")

	// Start input polling and resize handling
	go s.pollInput()
	go s.handleResize()

	return s, nil
}

// SetScrollRegion sets the scrolling region
func (s *Screen) SetScrollRegion(top, bottom int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scrollRegion.top = top
	s.scrollRegion.bottom = bottom
	fmt.Fprintf(s.out, "\x1b[%d;%dr", top+1, bottom+1)
}

// DisableScrollRegion disables the scrolling region
func (s *Screen) DisableScrollRegion() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scrollRegion.top = 0
	s.scrollRegion.bottom = s.height
	fmt.Fprintf(s.out, "\x1b[r")
}

// handleResize handles terminal resize events
func (s *Screen) handleResize() {
	for {
		select {
		case <-s.done:
			return
		case <-s.sigwinch:
			if width, height, err := term.GetSize(int(s.in.Fd())); err == nil {
				s.mu.Lock()
				s.width = width
				s.height = height
				s.mu.Unlock()
				s.events <- Key{Type: KeyTypeResize}
			}
		}
	}
}

// pollInput reads raw input and converts to Key events
func (s *Screen) pollInput() {
	buf := make([]byte, 32)
	
	for {
		select {
		case <-s.done:
			return
		default:
			n, err := s.in.Read(buf)
			if err != nil {
				continue
			}

			// Process escape sequences
			if n > 0 {
				switch {
				case buf[0] == 3: // Ctrl-C
					s.events <- Key{Type: KeyTypeSpecial, Special: KeyCtrlC}
				case buf[0] == '\r': // Enter
					s.events <- Key{Type: KeyTypeSpecial, Special: KeyEnter}
				case buf[0] == 127: // Backspace
					s.events <- Key{Type: KeyTypeSpecial, Special: KeyBackspace}
				case buf[0] == 27 && n > 2: // Escape sequence
					if buf[1] == '[' {
						switch buf[2] {
						case 'A':
							s.events <- Key{Type: KeyTypeSpecial, Special: KeyUp}
						case 'B':
							s.events <- Key{Type: KeyTypeSpecial, Special: KeyDown}
						case 'C':
							s.events <- Key{Type: KeyTypeSpecial, Special: KeyRight}
						case 'D':
							s.events <- Key{Type: KeyTypeSpecial, Special: KeyLeft}
						case 'H':
							s.events <- Key{Type: KeyTypeSpecial, Special: KeyHome}
						case 'F':
							s.events <- Key{Type: KeyTypeSpecial, Special: KeyEnd}
						case '5':
							if n > 3 && buf[3] == '~' {
								s.events <- Key{Type: KeyTypeSpecial, Special: KeyPgUp}
							}
						case '6':
							if n > 3 && buf[3] == '~' {
								s.events <- Key{Type: KeyTypeSpecial, Special: KeyPgDn}
							}
						case '3':
							if n > 3 && buf[3] == '~' {
								s.events <- Key{Type: KeyTypeSpecial, Special: KeyDelete}
							}
						}
					}
				default:
					if r, size := utf8.DecodeRune(buf[:n]); size > 0 && r != utf8.RuneError {
						s.events <- Key{Type: KeyTypeRune, Rune: r}
					}
				}
			}
		}
	}
}

// PollEvent returns the next input event
func (s *Screen) PollEvent() Key {
	return <-s.events
}

// SaveCursor saves the current cursor position
func (s *Screen) SaveCursor() {
	fmt.Fprint(s.out, "\x1b7")
}

// RestoreCursor restores the saved cursor position
func (s *Screen) RestoreCursor() {
	fmt.Fprint(s.out, "\x1b8")
}

// Write writes raw text directly to the terminal
func (s *Screen) Write(text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := fmt.Fprint(s.out, text)
	return err
}

// WriteAt writes text at a specific position
func (s *Screen) WriteAt(text string, x, y int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SaveCursor()
	_, err := fmt.Fprintf(s.out, "\x1b[%d;%dH%s", y+1, x+1, text)
	s.RestoreCursor()
	return err
}

// ClearLine clears the current line
func (s *Screen) ClearLine() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := fmt.Fprint(s.out, "\x1b[2K")
	return err
}

// ClearToEndOfLine clears from cursor to end of line
func (s *Screen) ClearToEndOfLine() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := fmt.Fprint(s.out, "\x1b[K")
	return err
}

// Clear clears the screen and resets cursor
func (s *Screen) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := fmt.Fprint(s.out, "\x1b[2J\x1b[H")
	return err
}

// ClearRegion clears a specific region of the screen
func (s *Screen) ClearRegion(startY, endY int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SaveCursor()
	for y := startY; y <= endY; y++ {
		fmt.Fprintf(s.out, "\x1b[%d;1H\x1b[2K", y+1)
	}
	s.RestoreCursor()
	return nil
}

// SetCursor sets the cursor position
func (s *Screen) SetCursor(x, y int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := fmt.Fprintf(s.out, "\x1b[%d;%dH", y+1, x+1)
	return err
}

// ShowCursor shows or hides the cursor
func (s *Screen) ShowCursor(show bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if show {
		_, err := fmt.Fprint(s.out, "\x1b[?25h")
		return err
	}
	_, err := fmt.Fprint(s.out, "\x1b[?25l")
	return err
}

// Size returns the terminal dimensions
func (s *Screen) Size() (width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.width, s.height
}

// Close cleans up the screen
func (s *Screen) Close() {
	signal.Stop(s.sigwinch)
	close(s.done)
	
	// Restore terminal state
	if s.oldState != nil {
		term.Restore(int(s.in.Fd()), s.oldState)
	}

	// Show cursor and switch back to main screen
	fmt.Fprint(s.out, "\x1b[?25h\x1b[?1049l")
	
	s.in.Close()
}
