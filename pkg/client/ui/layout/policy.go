package layout

// Policy defines how components should be laid out
type Policy interface {
	// ComputeLayout returns the regions for each component type
	ComputeLayout(width, height int) map[ComponentType]Region
}

// StandardPolicy implements a standard terminal layout:
// - Status bar at top (1 line)
// - Viewport in middle (variable height)
// - Input bar at bottom (1 line)
type StandardPolicy struct{}

// NewStandardPolicy creates a new standard layout policy
func NewStandardPolicy() *StandardPolicy {
	return &StandardPolicy{}
}

// ComputeLayout implements Policy interface
func (p *StandardPolicy) ComputeLayout(width, height int) map[ComponentType]Region {
	if height < 3 { // Minimum height needed for all components
		height = 3
	}
	if width < 1 {
		width = 1
	}

	return map[ComponentType]Region{
		StatusBarType: {
			Row:    0,
			Col:    0,
			Width:  width,
			Height: 1,
		},
		ViewportType: {
			Row:    1,
			Col:    0,
			Width:  width,
			Height: height - 2, // Space for status and input bars
		},
		InputBarType: {
			Row:    height - 1,
			Col:    0,
			Width:  width,
			Height: 1,
		},
	}
}
