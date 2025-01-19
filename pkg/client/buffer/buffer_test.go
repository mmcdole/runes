package buffer

import (
	"strings"
	"testing"
)

func TestUnterminatedColorPropagation(t *testing.T) {
	buf := New()

	// Write lines with unterminated color
	input := []string{
		"\033[31mLine 1", // Start red, unterminated
		"Line 2",         // Should inherit red
		"Line 3",         // Should inherit red
		"Line 4\033[0m",  // Should start red, then reset
	}

	// Write each line separately to simulate real input
	for _, line := range input {
		buf.Write(line + "\n")
	}

	// Get all lines
	lines := buf.GetLines(0, buf.Len())

	// Expected content for each line
	expected := []string{
		"\033[31mLine 1",      // Original red start
		"\033[31mLine 2",      // Should have red prepended
		"\033[31mLine 3",      // Should have red prepended
		"\033[31mLine 4\033[0m", // Should have red prepended, then reset
	}

	// Expected raw content (without ANSI state propagation)
	expectedRaw := []string{
		"\033[31mLine 1",
		"Line 2",
		"Line 3",
		"Line 4\033[0m",
	}

	// Verify each line
	for i, line := range lines {
		if line.Content != expected[i] {
			t.Errorf("Line %d content mismatch:\nExpected: %q\nGot: %q", 
				i+1, 
				strings.ReplaceAll(expected[i], "\033", "\\033"),
				strings.ReplaceAll(line.Content, "\033", "\\033"),
			)
		}
		if line.Raw != expectedRaw[i] {
			t.Errorf("Line %d raw content mismatch:\nExpected: %q\nGot: %q",
				i+1,
				strings.ReplaceAll(expectedRaw[i], "\033", "\\033"),
				strings.ReplaceAll(line.Raw, "\033", "\\033"),
			)
		}
	}
}
