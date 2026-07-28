package upgrade

import "idle-engine/internal/domain/modifier"

// Upgrade mewakili item peningkatan yang dapat dibeli oleh pemain.
type Upgrade struct {
	ID               string
	Name             string
	Description      string
	Cost             float64
	TargetBusinessID string
	Effect           modifier.Modifier
	IsPurchased      bool
}

// NewUpgrade membuat instance Upgrade baru.
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

// Purchase menandai upgrade telah dibeli.
func (u *Upgrade) Purchase() {
	u.IsPurchased = true
}
