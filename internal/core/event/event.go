package event

// Achievement represents a game achievement that awards bonuses.
type Achievement struct {
	ID               string
	Name             string
	Description      string
	ConditionType    string  // "balance" (for wallet balance), "level" (for business level)
	TargetBusinessID string  // Target business ID (e.g. "lemonade") or "" for global
	TargetValue      float64 // Level or balance threshold to unlock the achievement
	IsUnlocked       bool
	BonusMultiplier  float64 // Revenue bonus multiplier (e.g. 0.10 for +10% revenue)
}

// NewAchievement creates a new Achievement instance.
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

// Unlock unlocks the achievement.
func (a *Achievement) Unlock() {
	a.IsUnlocked = true
}

// CheckCondition validates whether achievement conditions are met.
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
