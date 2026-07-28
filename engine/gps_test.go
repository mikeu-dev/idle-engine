package engine

import (
	"testing"
)

func TestGetGPS(t *testing.T) {
	// Setup engine
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Ambil Lemonade Stand (idx 0)
	lemonade := e.GetBusinesses()[0]

	// Jika level 0, GPS harus 0
	if lemonade.GetGPS() != 0.0 {
		t.Errorf("expected level 0 GPS to be 0, got %f", lemonade.GetGPS())
	}

	// Upgrade ke level 1: cost = 4.0, BaseIncome = 1.0, Duration = 1.0s, SpeedMultiplier = 1.0
	// Maka GPS level 1 = 1.0 Poin/detik
	lemonade.Upgrade()
	if lemonade.GetGPS() != 1.0 {
		t.Errorf("expected level 1 GPS to be 1.0, got %f", lemonade.GetGPS())
	}

	// Upgrade ke level 2 -> GPS = 2.0 Poin/detik
	lemonade.Upgrade()
	if lemonade.GetGPS() != 2.0 {
		t.Errorf("expected level 2 GPS to be 2.0, got %f", lemonade.GetGPS())
	}
}

func TestGetTotalGPS(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Lini bisnis di index 1 (Newspaper Route) di-setup otomatis bawaan (IsAutomated = true)
	news := e.GetBusinesses()[1]
	// Naikkan level ke 1 -> BaseIncome = 4.0, Duration = 3.0s, maka GPS = 4 / 3 = 1.3333333333333333
	news.Upgrade()

	expectedGPS := 4.0 / 3.0
	diff := e.GetTotalGPS() - expectedGPS
	if diff < -1e-9 || diff > 1e-9 {
		t.Errorf("expected total GPS to be around %f, got %f", expectedGPS, e.GetTotalGPS())
	}

	// Lini bisnis index 0 (Lemonade) tidak otomatis, jadi level 1-nya tidak masuk hitungan Total GPS
	lemon := e.GetBusinesses()[0]
	lemon.Upgrade()
	diff = e.GetTotalGPS() - expectedGPS
	if diff < -1e-9 || diff > 1e-9 {
		t.Errorf("expected total GPS to still be around %f after manual business upgrade, got %f", expectedGPS, e.GetTotalGPS())
	}
}
