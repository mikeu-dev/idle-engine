package event

import "testing"

func TestAchievementConditions(t *testing.T) {
	// Buat achievement balance
	achBalance := NewAchievement("pts_1000", "Points Tycoon", "Collect 1,000 points", "balance", "", 1000.0, 0.10)

	// Buat achievement level
	achLevel := NewAchievement("lemon_10", "Lemonade Tycoon", "Lemonade Stand Level 10", "level", "lemonade", 10.0, 0.10)

	// 1. Kondisi awal: Keduanya belum terbuka
	if achBalance.IsUnlocked || achLevel.IsUnlocked {
		t.Error("expected achievements to be locked initially")
	}

	levels := map[string]int{
		"lemonade": 5,
	}

	// 2. Cek dengan saldo 500 dan level 5 (Keduanya harus tidak terpenuhi)
	if achBalance.CheckCondition(500.0, levels) {
		t.Error("expected balance achievement check to return false")
	}
	if achLevel.CheckCondition(500.0, levels) {
		t.Error("expected level achievement check to return false")
	}

	// 3. Naikkan saldo jadi 1500 (achBalance harus terpenuhi)
	if !achBalance.CheckCondition(1500.0, levels) {
		t.Error("expected balance achievement check to return true")
	}
	// Buka manual
	achBalance.Unlock()
	if !achBalance.IsUnlocked {
		t.Error("expected balance achievement to be unlocked")
	}

	// Cek ulang (harus mengembalikan false karena sudah terbuka)
	if achBalance.CheckCondition(1500.0, levels) {
		t.Error("expected CheckCondition to return false for already unlocked achievements")
	}

	// 4. Naikkan level bisnis jadi 10 (achLevel harus terpenuhi)
	levels["lemonade"] = 10
	if !achLevel.CheckCondition(500.0, levels) {
		t.Error("expected level achievement check to return true")
	}
}
