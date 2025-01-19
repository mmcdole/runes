package layout

// Component represents a UI component that can be rendered
type Component interface {
	// Render renders the component within its assigned region
	Render(Region)
}

// Manager handles the layout of UI components
type Manager struct {
	policy     Policy
	components map[ComponentType]Component
}

// NewManager creates a new layout manager
func NewManager(policy Policy) *Manager {
	return &Manager{
		policy:     policy,
		components: make(map[ComponentType]Component),
	}
}

// RegisterComponent adds a component to be managed
func (m *Manager) RegisterComponent(typ ComponentType, component Component) {
	m.components[typ] = component
}

// HandleResize updates the layout of all components
func (m *Manager) HandleResize(width, height int) {
	// Compute new layout
	regions := m.policy.ComputeLayout(width, height)

	// Update each component
	for typ, component := range m.components {
		if region, ok := regions[typ]; ok {
			component.Render(region)
		}
	}
}

// RenderAll renders all components in their current regions
func (m *Manager) RenderAll(width, height int) {
	m.HandleResize(width, height)
}
