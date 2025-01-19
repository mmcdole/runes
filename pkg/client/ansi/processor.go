package ansi

import (
	"fmt"
	"strings"
)

// colorState represents the current ANSI color and attribute state
type colorState struct {
	bold      bool
	dim       bool
	italic    bool
	underline bool
	blink     bool
	reverse   bool
	hidden    bool
	fgColor   int // -1 for default, 0-255 for extended colors
	bgColor   int // -1 for default, 0-255 for extended colors
}

// newColorState creates a new color state with default values
func newColorState() *colorState {
	return &colorState{
		fgColor: -1,
		bgColor: -1,
	}
}

// clone creates a copy of the color state
func (c *colorState) clone() *colorState {
	clone := *c
	return &clone
}

// reset resets all attributes to default
func (c *colorState) reset() {
	c.bold = false
	c.dim = false
	c.italic = false
	c.underline = false
	c.blink = false
	c.reverse = false
	c.hidden = false
	c.fgColor = -1
	c.bgColor = -1
}

// hasColor returns true if the state has any color or attribute set
func (c *colorState) hasColor() bool {
	return c.bold || c.dim || c.italic || c.underline || c.blink || 
	       c.reverse || c.hidden || c.fgColor >= 0 || c.bgColor >= 0
}

// toANSI converts the color state to an ANSI escape sequence
func (c *colorState) toANSI() string {
	if !c.hasColor() {
		return ""
	}

	var attrs []string
	
	// Add attributes in a consistent order
	if c.bold {
		attrs = append(attrs, "1")
	}
	if c.dim {
		attrs = append(attrs, "2")
	}
	if c.italic {
		attrs = append(attrs, "3")
	}
	if c.underline {
		attrs = append(attrs, "4")
	}
	if c.blink {
		attrs = append(attrs, "5")
	}
	if c.reverse {
		attrs = append(attrs, "7")
	}
	if c.hidden {
		attrs = append(attrs, "8")
	}
	if c.fgColor >= 0 {
		if c.fgColor < 8 {
			attrs = append(attrs, fmt.Sprintf("%d", c.fgColor+30))
		} else {
			attrs = append(attrs, fmt.Sprintf("38;5;%d", c.fgColor))
		}
	}
	if c.bgColor >= 0 {
		if c.bgColor < 8 {
			attrs = append(attrs, fmt.Sprintf("%d", c.bgColor+40))
		} else {
			attrs = append(attrs, fmt.Sprintf("48;5;%d", c.bgColor))
		}
	}

	return fmt.Sprintf("\033[%sm", strings.Join(attrs, ";"))
}

// processANSI processes ANSI escape sequences in a string and updates the color state
func (c *colorState) processANSI(str string) {
	// If the string doesn't start with ESC[, treat it as a raw sequence
	if !strings.HasPrefix(str, "\033[") {
		parts := strings.Split(str, ";")
		for j := 0; j < len(parts); j++ {
			part := parts[j]
			switch part {
			case "1":
				c.bold = true
			case "2":
				c.dim = true
			case "3":
				c.italic = true
			case "4":
				c.underline = true
			case "5":
				c.blink = true
			case "7":
				c.reverse = true
			case "8":
				c.hidden = true
			case "21", "22":
				c.bold = false
				c.dim = false
			case "23":
				c.italic = false
			case "24":
				c.underline = false
			case "25":
				c.blink = false
			case "27":
				c.reverse = false
			case "28":
				c.hidden = false
			case "30", "31", "32", "33", "34", "35", "36", "37":
				c.fgColor = atoi(part) - 30
			case "40", "41", "42", "43", "44", "45", "46", "47":
				c.bgColor = atoi(part) - 40
			case "38":
				if j+2 < len(parts) && parts[j+1] == "5" {
					if color := atoi(parts[j+2]); color >= 0 {
						c.fgColor = color
					}
					j += 2
				}
			case "48":
				if j+2 < len(parts) && parts[j+1] == "5" {
					if color := atoi(parts[j+2]); color >= 0 {
						c.bgColor = color
					}
					j += 2
				}
			case "39":
				c.fgColor = -1
			case "49":
				c.bgColor = -1
			}
		}
		return
	}

	// Process full ANSI sequences
	for i := 0; i < len(str); i++ {
		if str[i] == '\033' && i+1 < len(str) && str[i+1] == '[' {
			// Find the end of the sequence
			end := strings.IndexByte(str[i:], 'm')
			if end == -1 {
				continue
			}
			end += i

			// Parse the sequence
			seq := str[i+2 : end]
			if seq == "0" || seq == "" {
				c.reset()
				continue
			}

			// Process the sequence without the ESC[ prefix
			c.processANSI(seq)
			i = end
		}
	}
}

// Helper function to convert string to int
func atoi(s string) int {
	var n int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		} else {
			return -1
		}
	}
	return n
}
