package terminal

// Key represents a terminal key code
type Key uint16

// Modifier represents key modifiers (ctrl, alt, etc)
type Modifier uint8

// Modifier constants
const (
	ModNone Modifier = 0
	ModCtrl Modifier = 1 << iota
	ModAlt
	ModShift
)

// Key constants
const (
	KeyUnknown Key = iota
	KeyCtrlC
	KeyEnter
	KeyTab
	KeyBackspace
	KeyEscape
	KeyDelete
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// Internal ANSI sequence constants
const (
	esc      = 27    // \033 or \x1b
	csi      = '['   // Control Sequence Introducer
	tilde    = '~'   // Used by some special keys
	del      = '3'   // Delete key code
	pageUp   = '5'   // PageUp key code
	pageDn   = '6'   // PageDown key code
	arrowUp  = 'A'   // Up arrow
	arrowDn  = 'B'   // Down arrow
	arrowRt  = 'C'   // Right arrow
	arrowLt  = 'D'   // Left arrow
)

// ParseKey parses a raw byte sequence into a Key and Modifier
func ParseKey(input []byte) (Key, Modifier) {
	// Handle simple keys
	if len(input) == 1 {
		switch input[0] {
		case 3:
			return KeyCtrlC, ModCtrl
		case '\r':
			return KeyEnter, ModNone
		case '\t':
			return KeyTab, ModNone
		case 127:
			return KeyBackspace, ModNone
		}
		return KeyUnknown, ModNone
	}

	// Handle escape sequences
	if len(input) >= 3 && input[0] == esc && input[1] == csi {
		switch input[2] {
		case arrowUp:
			return KeyUp, ModNone
		case arrowDn:
			return KeyDown, ModNone
		case arrowRt:
			return KeyRight, ModNone
		case arrowLt:
			return KeyLeft, ModNone
		}

		// Handle extended sequences
		if len(input) == 4 && input[3] == tilde {
			switch input[2] {
			case pageUp:
				return KeyPageUp, ModNone
			case pageDn:
				return KeyPageDown, ModNone
			case del:
				return KeyDelete, ModNone
			}
		}
	}

	return KeyUnknown, ModNone
}

// Key detection helpers
func IsPageUp(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyPageUp
}

func IsPageDown(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyPageDown
}

func IsUpArrow(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyUp
}

func IsDownArrow(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyDown
}

func IsLeftArrow(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyLeft
}

func IsRightArrow(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyRight
}

func IsCtrlC(input []byte) bool {
	key, mod := ParseKey(input)
	return key == KeyCtrlC && mod == ModCtrl
}

func IsBackspace(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyBackspace
}

func IsEnter(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyEnter
}

func IsDelete(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyDelete
}

func IsHome(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyHome
}

func IsEnd(input []byte) bool {
	key, _ := ParseKey(input)
	return key == KeyEnd
}
