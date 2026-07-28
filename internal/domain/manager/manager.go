package manager

// Manager merepresentasikan manajer yang dapat dipekerjakan untuk otomatisasi.
type Manager struct {
	ID               string
	Name             string
	Description      string
	Cost             float64
	TargetBusinessID string
	IsHired          bool
}

// NewManager membuat instance Manager baru.
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

// Hire menandai manajer telah dipekerjakan.
func (m *Manager) Hire() {
	m.IsHired = true
}
