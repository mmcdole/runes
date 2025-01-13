package screen

import (
	"fmt"
	"os"
	"sync"
	"unicode/utf8"

	"golang.org/x/term"
)

// Key represents a keyboard input
type Key struct {
	Type    KeyType
	Rune    rune
	Special SpecialKey
}

type KeyType int

const (
	KeyTypeRune KeyType = iota
	KeyTypeSpecial
)

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
	in       *os.File    // Terminal input
	out      *os.File    // Terminal output
	oldState *term.State // Original terminal state
	width    int
	height   int
	mu       sync.Mutex
	events   chan Key
	done     chan struct{}
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
	}

	// Switch to alternate screen and hide cursor
	fmt.Fprint(s.out, "\x1b[?1049h\x1b[?25l")

	// Start input polling
	go s.pollInput()

	return s, nil
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
	_, err := fmt.Fprintf(s.out, "\x1b[%d;%dH%s", y+1, x+1, text)
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

	// Save cursor
	fmt.Fprint(s.out, "\x1b7")

	// Clear each line in region
	for y := startY; y <= endY; y++ {
		fmt.Fprintf(s.out, "\x1b[%d;1H\x1b[2K", y+1)
	}

	// Restore cursor
	fmt.Fprint(s.out, "\x1b8")
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
	close(s.done)
	
	// Restore terminal state
	if s.oldState != nil {
		term.Restore(int(s.in.Fd()), s.oldState)
	}

	// Show cursor and switch back to main screen
	fmt.Fprint(s.out, "\x1b[?25h\x1b[?1049l")
	
	s.in.Close()
}
