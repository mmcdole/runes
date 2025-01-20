package ansi

import (
	"strings"
	"testing"
)

func TestLineProcessor_ColorPropagation(t *testing.T) {
	processor := NewLineProcessor()

	input := []string{
		"\033[31mLine 1",      // Set red
		"Line 2",              // Should inherit red
		"Line 3",              // Should inherit red
		"Line 4\033[0m",       // Should inherit red, then reset
		"\033[32mLine 5",      // Set green
		"\033[1mLine 6",       // Add bold to green
		"Line 7\033[0m",       // Should inherit green+bold
	}

	expected := []string{
		"\033[31mLine 1\033[0m",
		"\033[31mLine 2\033[0m",
		"\033[31mLine 3\033[0m",
		"\033[31mLine 4\033[0m",
		"\033[32mLine 5\033[0m",
		"\033[32m\033[1mLine 6\033[0m",
		"\033[32;1mLine 7\033[0m",
	}

	for i, content := range input {
		line := processor.ProcessLine(content)
		if line.Content != expected[i] {
			t.Errorf("Line %d content mismatch:\nExpected: %q\nGot: %q", 
				i+1, 
				strings.ReplaceAll(expected[i], "\033", "\\033"),
				strings.ReplaceAll(line.Content, "\033", "\\033"),
			)
		}

		// Verify raw content is preserved
		if line.Raw != input[i] {
			t.Errorf("Line %d raw content mismatch:\nExpected: %q\nGot: %q",
				i+1,
				strings.ReplaceAll(input[i], "\033", "\\033"),
				strings.ReplaceAll(line.Raw, "\033", "\\033"),
			)
		}
	}
}

func TestLineProcessor_ComplexColorStates(t *testing.T) {
	processor := NewLineProcessor()

	input := []string{
		"\033[31;1mBold Red",      // Sets red+bold
		"\033[22mNot Bold Red",    // Removes bold
		"\033[32mGreen\033[1mBold", // Sets green then bold
		"\033[39mDefault Bold",    // Removes color
		"\033[0;34mBlue\033[0m",   // Sets blue
	}

	expected := []string{
		// No previous state
		"\033[31;1mBold Red\033[0m",

		// Previous state is red+bold
		"\033[31;1m\033[22mNot Bold Red\033[0m",

		// Previous state is red (not bold)
		"\033[31m\033[32mGreen\033[1mBold\033[0m",

		// Previous state is green+bold
		"\033[32;1m\033[39mDefault Bold\033[0m",

		// Previous state is bold (no color)
		"\033[1m\033[0;34mBlue\033[0m",
	}

	for i, content := range input {
		line := processor.ProcessLine(content)
		if line.Content != expected[i] {
			t.Errorf("Line %d content mismatch:\nExpected: %q\nGot: %q", 
				i+1, 
				strings.ReplaceAll(expected[i], "\033", "\\033"),
				strings.ReplaceAll(line.Content, "\033", "\\033"),
			)
		}
	}
}
