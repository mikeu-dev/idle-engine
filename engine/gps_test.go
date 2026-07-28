package engine

import (
	"testing"
)

func TestGetGPS(t *testing.T) {
	// Setup engine
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Retrieve Lemonade Stand (idx 0)
	lemonade := e.GetBusinesses()[0]

	// If level is 0, GPS must be 0
	if lemonade.GetGPS() != 0.0 {
		t.Errorf("expected level 0 GPS to be 0, got %f", lemonade.GetGPS())
	}

	// Upgrade to level 1: cost = 4.0, BaseIncome = 1.0, Duration = 1.0s, SpeedMultiplier = 1.0
	// Hence, GPS level 1 = 1.0 points/second
	lemonade.Upgrade()
	if lemonade.GetGPS() != 1.0 {
		t.Errorf("expected level 1 GPS to be 1.0, got %f", lemonade.GetGPS())
	}

	// Upgrade to level 2 -> GPS = 2.0 points/second
	lemonade.Upgrade()
	if lemonade.GetGPS() != 2.0 {
		t.Errorf("expected level 2 GPS to be 2.0, got %f", lemonade.GetGPS())
	}
}

func TestGetTotalGPS(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Business line at index 1 (Newspaper Route) is automated by default (IsAutomated = true)
	news := e.GetBusinesses()[1]
	// Upgrade to level 1 -> BaseIncome = 4.0, Duration = 3.0s, GPS = 4 / 3 = 1.333333
	news.Upgrade()

	expectedGPS := 4.0 / 3.0
	diff := e.GetTotalGPS() - expectedGPS
	if diff < -1e-9 || diff > 1e-9 {
		t.Errorf("expected total GPS to be around %f, got %f", expectedGPS, e.GetTotalGPS())
	}

	// Business at index 0 (Lemonade) is manual, so its level 1 does not count towards Total GPS
	lemon := e.GetBusinesses()[0]
	lemon.Upgrade()
	diff = e.GetTotalGPS() - expectedGPS
	if diff < -1e-9 || diff > 1e-9 {
		t.Errorf("expected total GPS to still be around %f after manual business upgrade, got %f", expectedGPS, e.GetTotalGPS())
	}
}
