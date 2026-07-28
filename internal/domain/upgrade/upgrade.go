package upgrade

import "idle-engine/internal/domain/modifier"

// Upgrade represents an upgrade item that can be purchased by the player.
type Upgrade struct {
	ID               string
	Name             string
	Description      string
	Cost             float64
	TargetBusinessID string
	Effect           modifier.Modifier
	IsPurchased      bool
}

// NewUpgrade creates a new Upgrade instance.
func NewUpgrade(id, name, desc string, cost float64, target string, effect modifier.Modifier) *Upgrade {
	return &Upgrade{
		ID:               id,
		Name:             name,
		Description:      desc,
		Cost:             cost,
		TargetBusinessID: target,
		Effect:           effect,
		IsPurchased:      false,
	}
}

// Purchase marks the upgrade as purchased.
func (u *Upgrade) Purchase() {
	u.IsPurchased = true
}
