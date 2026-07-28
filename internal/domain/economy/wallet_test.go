package economy

import (
	"testing"
)

func TestWallet(t *testing.T) {
	w := NewWallet(100.0)

	// Test initial balance
	if w.Balance() != 100.0 {
		t.Errorf("expected balance 100.0, got %f", w.Balance())
	}

	// Test Add
	w.Add(50.0)
	if w.Balance() != 150.0 {
		t.Errorf("expected balance 150.0 after add, got %f", w.Balance())
	}

	// Test Add negative/zero (should be ignored)
	w.Add(-10.0)
	if w.Balance() != 150.0 {
		t.Errorf("expected negative add to be ignored, got %f", w.Balance())
	}

	// Test CanAfford
	if !w.CanAfford(150.0) {
		t.Error("expected to afford 150.0")
	}
	if w.CanAfford(150.01) {
		t.Error("expected not to afford 150.01")
	}

	// Test Spend success
	success := w.Spend(50.0)
	if !success {
		t.Error("expected spend to succeed")
	}
	if w.Balance() != 100.0 {
		t.Errorf("expected balance 100.0 after spend, got %f", w.Balance())
	}

	// Test Spend insufficient
	success = w.Spend(150.0)
	if success {
		t.Error("expected spend to fail due to insufficient funds")
	}
	if w.Balance() != 100.0 {
		t.Errorf("expected balance to remain 100.0, got %f", w.Balance())
	}

	// Test Spend negative/zero
	success = w.Spend(-5.0)
	if success {
		t.Error("expected negative spend to fail")
	}
}
