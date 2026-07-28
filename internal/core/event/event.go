package event

// Achievement mewakili pencapaian di dalam game yang memberikan bonus.
type Achievement struct {
	ID               string
	Name             string
	Description      string
	ConditionType    string  // "balance" (untuk saldo), "level" (untuk tingkat level bisnis)
	TargetBusinessID string  // ID bisnis target (misal "lemonade") atau "" untuk global
	TargetValue      float64 // Batas level atau saldo untuk membuka pencapaian
	IsUnlocked       bool
	BonusMultiplier  float64 // Bonus multiplier pendapatan (misal 0.10 untuk +10% pendapatan)
}

// NewAchievement membuat instance Achievement baru.
func NewAchievement(id, name, desc, condType, target string, val, bonus float64) *Achievement {
	return &Achievement{
		ID:               id,
		Name:             name,
		Description:      desc,
		ConditionType:    condType,
		TargetBusinessID: target,
		TargetValue:      val,
		IsUnlocked:       false,
		BonusMultiplier:  bonus,
	}
}

// Unlock membuka pencapaian.
func (a *Achievement) Unlock() {
	a.IsUnlocked = true
}

// CheckCondition memvalidasi apakah kondisi pencapaian terpenuhi.
func (a *Achievement) CheckCondition(balance float64, businessLevels map[string]int) bool {
	if a.IsUnlocked {
		return false
	}

	switch a.ConditionType {
	case "balance":
		return balance >= a.TargetValue
	case "level":
		if lv, ok := businessLevels[a.TargetBusinessID]; ok {
			return float64(lv) >= a.TargetValue
		}
	}
	return false
}
