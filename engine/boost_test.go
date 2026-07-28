package engine

import (
	"testing"
	"time"
)

func TestTriggerSuperBoost(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Add balance (initial 4.0 + 100.0 = 104.0)
	e.wallet.Add(100.0)

	// Buy Lemonade Stand (cost 4.0, balance left = 100.0)
	e.BuyUpgrade(0)

	// Ensure initial boost duration is 0
	if e.GetBoostDuration() != 0 {
		t.Errorf("expected initial boost duration to be 0, got %v", e.GetBoostDuration())
	}

	// Trigger Super Boost (cost 50.0, balance left = 50.0)
	success := e.TriggerSuperBoost()
	if !success {
		t.Fatal("expected Super Boost trigger to succeed")
	}

	// Verify wallet balance is deducted (104 - 4 - 50 = 50.0)
	if e.GetWallet().Balance() != 50.0 {
		t.Errorf("expected wallet balance 50.0, got %f", e.GetWallet().Balance())
	}

	// Verify remaining boost duration is 30 seconds
	if e.GetBoostDuration() != 30*time.Second {
		t.Errorf("expected boost duration to be 30s, got %v", e.GetBoostDuration())
	}

	// Verify that Update elapsed time decreases the boost duration
	e.lastTick = time.Now()
	time.Sleep(10 * time.Millisecond)
	e.Update()

	if e.GetBoostDuration() >= 30*time.Second || e.GetBoostDuration() <= 0 {
		t.Errorf("expected boost duration to decrease, got %v", e.GetBoostDuration())
	}
}

func TestTriggerTimeWarp(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Add balance (initial 4.0 + 200.0 = 204.0)
	// Note: Adding balance > 100.0 automatically unlocks Achievement "pts_100" (+10% global revenue multiplier).
	e.wallet.Add(200.0)

	// Buy Newspaper upgrade manually via engine to deduct balance (cost 20.0, balance left = 184.0)
	successBuy := e.BuyUpgrade(1)
	if !successBuy {
		t.Fatal("expected to buy Newspaper upgrade successfully")
	}

	// Trigger Time Warp (cost 150.0, balance left = 34.0)
	success := e.TriggerTimeWarp()
	if !success {
		t.Fatal("expected Time Warp to succeed")
	}

	// Calculate expected balance:
	// Wallet balance after cost = 34.0
	// Newspaper income (Lv 1) = 4.0 * 1.1 (achievement bonus) = 4.4
	// Newspaper GPS = 4.4 / 3.0s = 1.466667 / second
	// Time Warp 1 Hour = 1.466667 * 3600 = 5280.0
	// Expected Final Balance = 34.0 + 5280.0 = 5314.0
	expectedBalance := 5314.0
	diff := e.GetWallet().Balance() - expectedBalance
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("expected wallet balance to be around %f, got %f", expectedBalance, e.GetWallet().Balance())
	}
}
