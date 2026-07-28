package event

import "testing"

func TestAchievementConditions(t *testing.T) {
		// Create balance achievement
	achBalance := NewAchievement("pts_1000", "Points Tycoon", "Collect 1,000 points", "balance", "", 1000.0, 0.10)

	// Create level achievement
	achLevel := NewAchievement("lemon_10", "Lemonade Tycoon", "Lemonade Stand Level 10", "level", "lemonade", 10.0, 0.10)

	// 1. Initial state: Both locked
	if achBalance.IsUnlocked || achLevel.IsUnlocked {
		t.Error("expected achievements to be locked initially")
	}

	levels := map[string]int{
		"lemonade": 5,
	}

	// 2. Check with balance 500 and level 5 (both should be unsatisfied)
	if achBalance.CheckCondition(500.0, levels) {
		t.Error("expected balance achievement check to return false")
	}
	if achLevel.CheckCondition(500.0, levels) {
		t.Error("expected level achievement check to return false")
	}

	// 3. Increase balance to 1500 (achBalance should be satisfied)
	if !achBalance.CheckCondition(1500.0, levels) {
		t.Error("expected balance achievement check to return true")
	}
	// Unlock manually
	achBalance.Unlock()
	if !achBalance.IsUnlocked {
		t.Error("expected balance achievement to be unlocked")
	}

	// Check again (should return false since it is already unlocked)
	if achBalance.CheckCondition(1500.0, levels) {
		t.Error("expected CheckCondition to return false for already unlocked achievements")
	}

	// 4. Increase business level to 10 (achLevel should be satisfied)
	levels["lemonade"] = 10
	if !achLevel.CheckCondition(500.0, levels) {
		t.Error("expected level achievement check to return true")
	}
}
