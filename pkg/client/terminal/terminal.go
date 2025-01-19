package terminal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Control sequences
const (
	seqAlternateScreen = "\033[?1049h"
	seqMainScreen     = "\033[?1049l"
	seqClearScreen    = "\033[2J"
	seqHome          = "\033[H"
	seqHideCursor    = "\033[?25l"
	seqShowCursor    = "\033[?25h"
	seqOriginMode    = "\033[?6h"
	seqClearLine     = "\033[2K"
)

// Common errors
var (
	ErrNotInitialized     = errors.New("terminal not initialized")
	ErrAlreadyInitialized = errors.New("terminal already initialized")
)

// Terminal represents the terminal interface
type Terminal struct {
	fd         int
	oldState   *term.State
	width      int
	height     int
	resizeChan chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
}

// New creates a new terminal instance
func New() *Terminal {
	ctx, cancel := context.WithCancel(context.Background())
	return &Terminal{
		fd:         int(os.Stdout.Fd()),
		resizeChan: make(chan struct{}, 1),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Init initializes the terminal in raw mode
func (t *Terminal) Init() error {
	if t.oldState != nil {
		return ErrAlreadyInitialized
	}

	var err error
	t.oldState, err = term.MakeRaw(t.fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}

	// Set up signal handling for window resize
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				signal.Stop(sigChan)
				return
			case <-sigChan:
				if err := t.updateSize(); err == nil {
					select {
					case t.resizeChan <- struct{}{}:
					default:
						// Channel full, skip
					}
				}
			}
		}
	}()

	if err := t.updateSize(); err != nil {
		t.Cleanup()
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	// Initialize terminal state
	if err := t.writeSeq([]byte(seqAlternateScreen)); err != nil {
		t.Cleanup()
		return fmt.Errorf("failed to enter alternate screen: %w", err)
	}
	if err := t.Clear(); err != nil {
		t.Cleanup()
		return fmt.Errorf("failed to clear screen: %w", err)
	}
	if err := t.writeSeq([]byte(seqHideCursor)); err != nil {
		t.Cleanup()
		return fmt.Errorf("failed to hide cursor: %w", err)
	}

	return nil
}

// writeSeq writes a control sequence to the terminal
func (t *Terminal) writeSeq(seq []byte) error {
	if err := t.checkState(); err != nil {
		return err
	}
	_, err := t.Write(seq)
	return err
}

// checkState validates terminal state
func (t *Terminal) checkState() error {
	if t.oldState == nil {
		return ErrNotInitialized
	}
	return nil
}

// updateSize updates the stored terminal dimensions
func (t *Terminal) updateSize() error {
	width, height, err := term.GetSize(t.fd)
	if err != nil {
		return err
	}
	t.width = width
	t.height = height
	return nil
}

// ResizeChan returns a channel that receives notifications when the terminal is resized
func (t *Terminal) ResizeChan() <-chan struct{} {
	return t.resizeChan
}

// Cleanup restores the terminal to its original state
func (t *Terminal) Cleanup() error {
	if t.oldState != nil {
		t.cancel() // Stop resize handler

		// Restore terminal state
		if err := t.writeSeq([]byte(seqShowCursor)); err != nil {
			return fmt.Errorf("failed to show cursor: %w", err)
		}
		if err := t.writeSeq([]byte(seqMainScreen)); err != nil {
			return fmt.Errorf("failed to restore main screen: %w", err)
		}
		if err := term.Restore(t.fd, t.oldState); err != nil {
			return fmt.Errorf("failed to restore terminal state: %w", err)
		}
		t.oldState = nil
	}
	return nil
}

// Size returns the current terminal dimensions
func (t *Terminal) Size() (width, height int) {
	return t.width, t.height
}

// Write writes data directly to the terminal
func (t *Terminal) Write(data []byte) (int, error) {
	if err := t.checkState(); err != nil {
		return 0, err
	}
	return os.Stdout.Write(data)
}

// Read reads data from the terminal
func (t *Terminal) Read(data []byte) (int, error) {
	if err := t.checkState(); err != nil {
		return 0, err
	}
	return os.Stdin.Read(data)
}

// MoveCursor moves the cursor to the specified position
func (t *Terminal) MoveCursor(row, col int) error {
	return t.writeSeq([]byte(fmt.Sprintf("\033[%d;%dH", row, col)))
}

// SetScrollRegion sets the scrollable region
func (t *Terminal) SetScrollRegion(top, bottom int) error {
	// Set margins and ensure we're in the scrolling region
	if err := t.writeSeq([]byte(fmt.Sprintf("\033[%d;%dr", top, bottom))); err != nil {
		return err
	}
	// Ensure we're using origin mode relative to scroll region
	if err := t.writeSeq([]byte(seqOriginMode)); err != nil {
		return err
	}
	// Move cursor to start of region
	return t.MoveCursor(top, 1)
}

// Clear clears the entire terminal screen
func (t *Terminal) Clear() error {
	if err := t.writeSeq([]byte(seqClearScreen)); err != nil {
		return err
	}
	return t.writeSeq([]byte(seqHome))
}

// ClearLine clears the current line from cursor position
func (t *Terminal) ClearLine() error {
	if err := t.checkState(); err != nil {
		return err
	}
	return t.writeSeq([]byte(seqClearLine))
}
