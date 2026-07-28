package engine

import (
	"testing"
)

func TestBuyUpgradeMax(t *testing.T) {
	// Buat engine baru dengan data fallback (Lemonade Stand cost = 4.0, multiplier = 1.15)
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Set saldo awal wallet agar cukup membeli beberapa level lemonade stand
	// Level 0 -> 1: cost = 4.0
	// Level 1 -> 2: cost = 4.60
	// Level 2 -> 3: cost = 5.29
	// Total untuk 3 level = 13.89 poin.
	e.wallet.Add(10.0) // Tambahkan 10 poin (saldo total = 14.0)

	// Lemonade stand ada di idx 0
	success := e.BuyUpgradeMax(0)
	if !success {
		t.Fatal("expected BuyUpgradeMax to succeed")
	}

	b := e.GetBusinesses()[0]
	// Harus naik ke level 3 (karena level awal 0, saldo 14 poin cukup untuk 3 level seharga 13.89)
	if b.GetLevel() != 3 {
		t.Errorf("expected lemonade stand level to be 3, got %d", b.GetLevel())
	}

	// Sisa saldo haruslah 14.0 - 13.89 = ~0.11
	expectedBalance := 14.0 - 13.89
	diff := e.GetWallet().Balance() - expectedBalance
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("expected balance to be around %f, got %f", expectedBalance, e.GetWallet().Balance())
	}
}
