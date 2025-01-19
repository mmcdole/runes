package layout

// Region represents a rectangular area in the terminal
type Region struct {
	Row    int
	Col    int
	Width  int
	Height int
}

// ComponentType identifies different UI components
type ComponentType int

const (
	StatusBarType ComponentType = iota
	ViewportType
	InputBarType
)

// String returns the string representation of ComponentType
func (t ComponentType) String() string {
	switch t {
	case StatusBarType:
		return "StatusBar"
	case ViewportType:
		return "Viewport"
	case InputBarType:
		return "InputBar"
	default:
		return "Unknown"
	}
}
