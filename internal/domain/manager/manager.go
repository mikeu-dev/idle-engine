package manager

// Manager represents a manager that can be hired for automation.
type Manager struct {
	ID               string
	Name             string
	Description      string
	Cost             float64
	TargetBusinessID string
	IsHired          bool
}

// NewManager creates a new Manager instance.
func NewManager(id, name, desc string, cost float64, target string) *Manager {
	return &Manager{
		ID:               id,
		Name:             name,
		Description:      desc,
		Cost:             cost,
		TargetBusinessID: target,
		IsHired:          false,
	}
}

// Hire marks the manager as hired.
func (m *Manager) Hire() {
	m.IsHired = true
}
