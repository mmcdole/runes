package buffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/mmcdole/runes/pkg/client/types"
	"github.com/stretchr/testify/assert"
)

func TestBuffer_UnterminatedLines(t *testing.T) {
	tests := []struct {
		name     string
		writes   []string
		wantLen  int
		wantLast types.Line
	}{
		{
			name:    "single unterminated line",
			writes:  []string{"What is your name: "},
			wantLen: 1,
			wantLast: types.Line{
				Raw:     "What is your name: ",
				Display: "What is your name: ",
			},
		},
		{
			name:    "unterminated line becomes terminated",
			writes:  []string{"What is ", "your name: \n"},
			wantLen: 1,
			wantLast: types.Line{
				Raw:     "What is your name: ",
				Display: "What is your name: ",
			},
		},
		{
			name:    "multiple writes with mix of terminated/unterminated",
			writes:  []string{"First line\n", "Second ", "line continues\n", "Third incomplete"},
			wantLen: 3,
			wantLast: types.Line{
				Raw:     "Third incomplete",
				Display: "Third incomplete",
			},
		},
		{
			name:    "ANSI codes preserved across writes",
			writes:  []string{"\x1b[31mRed ", "text continues\n"},
			wantLen: 1,
			wantLast: types.Line{
				Raw:     "\x1b[31mRed text continues",
				Display: "\x1b[31mRed text continues\x1b[0m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(100)
			for _, write := range tt.writes {
				b.Write(types.NewLine(write))
			}

			assert.Equal(t, tt.wantLen, b.Len())
			if tt.wantLen > 0 {
				lines := b.GetLines(0, b.Len())
				assert.Equal(t, tt.wantLast.Raw, lines[len(lines)-1].Raw)
				assert.Equal(t, tt.wantLast.Display, lines[len(lines)-1].Display)
			}
		})
	}
}

func TestBuffer_LineAccumulation(t *testing.T) {
	b := New(100)

	// Write some complete lines
	b.Write(types.NewLine("First line\n"))
	b.Write(types.NewLine("Second line\n"))
	
	// Write an incomplete line
	b.Write(types.NewLine("Third "))
	b.Write(types.NewLine("line"))
	b.Write(types.NewLine(" continues\n"))

	// Check line count
	assert.Equal(t, 3, b.Len())

	// Verify line contents
	lines := b.GetLines(0, b.Len())
	assert.Equal(t, "First line", lines[0].Raw)
	assert.Equal(t, "Second line", lines[1].Raw)
	assert.Equal(t, "Third line continues", lines[2].Raw)
}

func TestBuffer_WindowedRetrieval(t *testing.T) {
	b := New(100)

	// Add 10 lines
	for i := 0; i < 10; i++ {
		b.Write(types.NewLine(fmt.Sprintf("Line %d\n", i)))
	}

	// Test retrieving all lines
	lines := b.GetLines(0, b.Len())
	assert.Equal(t, 10, len(lines))
	for i, line := range lines {
		assert.Equal(t, fmt.Sprintf("Line %d", i), line.Raw)
	}

	// Test retrieving a window in the middle
	lines = b.GetLines(3, 7)
	assert.Equal(t, 4, len(lines))
	for i, line := range lines {
		assert.Equal(t, fmt.Sprintf("Line %d", i+3), line.Raw)
	}

	// Test retrieving window at start
	lines = b.GetLines(0, 3)
	assert.Equal(t, 3, len(lines))
	for i, line := range lines {
		assert.Equal(t, fmt.Sprintf("Line %d", i), line.Raw)
	}

	// Test retrieving window at end
	lines = b.GetLines(7, 10)
	assert.Equal(t, 3, len(lines))
	for i, line := range lines {
		assert.Equal(t, fmt.Sprintf("Line %d", i+7), line.Raw)
	}

	// Test invalid window (start > end)
	lines = b.GetLines(5, 2)
	assert.Equal(t, 0, len(lines))

	// Test invalid window (start < 0)
	lines = b.GetLines(-1, 5)
	assert.Equal(t, 5, len(lines))

	// Test invalid window (end > len)
	lines = b.GetLines(5, 15)
	assert.Equal(t, 5, len(lines))
}

func TestBuffer_ANSIStatePreservation(t *testing.T) {
	b := New(100)

	// Write lines with ANSI codes to test color propagation
	b.Write(types.NewLine("\x1b[31mRed text\n"))         // Start red, no reset
	b.Write(types.NewLine("Still red\n"))                // Should inherit red
	b.Write(types.NewLine("\x1b[32mGreen\x1b[0m\n"))    // Switch to green, then reset
	b.Write(types.NewLine("Back to normal\n"))           // No color after reset
	b.Write(types.NewLine("\x1b[34mBlue"))              // Switch to blue, no newline

	// Get lines and verify ANSI state
	lines := b.GetLines(0, b.Len())
	assert.Equal(t, 5, len(lines))

	// Raw should preserve exactly what was sent
	assert.Equal(t, "\x1b[31mRed text", lines[0].Raw)
	assert.Equal(t, "Still red", lines[1].Raw)
	assert.Equal(t, "\x1b[32mGreen\x1b[0m", lines[2].Raw)
	assert.Equal(t, "Back to normal", lines[3].Raw)
	assert.Equal(t, "\x1b[34mBlue", lines[4].Raw)

	// Display should show proper color propagation
	assert.Equal(t, "\x1b[31mRed text\x1b[0m", lines[0].Display)
	assert.Equal(t, "\x1b[31mStill red\x1b[0m", lines[1].Display)  // Inherits red
	assert.Equal(t, "\x1b[32mGreen\x1b[0m", lines[2].Display)      // Green with reset
	assert.Equal(t, "Back to normal", lines[3].Display)            // No color after reset
	assert.Equal(t, "\x1b[34mBlue\x1b[0m", lines[4].Display)      // Blue with reset
}
