package engine

import (
	"testing"
)

func TestBuyUpgradeMax(t *testing.T) {
	// Create new engine with fallback data (Lemonade Stand cost = 4.0, multiplier = 1.15)
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Set wallet balance so it is sufficient to buy multiple levels
	// Level 0 -> 1: cost = 4.0
	// Level 1 -> 2: cost = 4.60
	// Level 2 -> 3: cost = 5.29
	// Total for 3 levels = 13.89 points.
	e.wallet.Add(10.0) // Add 10 points (total balance = 14.0)

	// Lemonade stand is at index 0
	success := e.BuyUpgradeMax(0)
	if !success {
		t.Fatal("expected BuyUpgradeMax to succeed")
	}

	b := e.GetBusinesses()[0]
	// Level should increase to 3 (starting level 0, 14.0 balance is sufficient for 3 levels costing 13.89)
	if b.GetLevel() != 3 {
		t.Errorf("expected lemonade stand level to be 3, got %d", b.GetLevel())
	}

	// Remaining balance should be around 14.0 - 13.89 = ~0.11
	expectedBalance := 14.0 - 13.89
	diff := e.GetWallet().Balance() - expectedBalance
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("expected balance to be around %f, got %f", expectedBalance, e.GetWallet().Balance())
	}
}
