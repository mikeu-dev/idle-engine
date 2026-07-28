package engine

import (
	"testing"
)

func TestPrestigeSystem(t *testing.T) {
	e := NewEngine()

	// 1. Awalnya belum ada investor
	if e.GetAngels() != 0 {
		t.Errorf("expected 0 angels initially, got %d", e.GetAngels())
	}

	// 2. Set lifetime earnings agar bisa klaim investor (900 Poin)
	// Kita simulasikan dengan menambahkan penghasilan
	e.lifetimeEarnings = 900.0

	claimable := e.CalculateAngelsToClaim()
	if claimable != 3 { // floor(sqrt(900/100)) = floor(sqrt(9)) = 3
		t.Errorf("expected 3 claimable angels, got %d", claimable)
	}

	// Beli Lemon Pitcher upgrade untuk memverifikasi reset-nya nanti
	e.wallet.Add(100)
	if !e.BuyUpgradeCard(0) {
		t.Fatal("failed to buy Lemon Pitcher upgrade for test preparation")
	}

	if !e.upgrades[0].IsPurchased {
		t.Error("expected Lemon Pitcher to be purchased")
	}

	// 3. Picu Prestige
	if !e.ClaimPrestige() {
		t.Fatal("expected ClaimPrestige to return true")
	}

	// 4. Verifikasi setelah reset prestis
	if e.GetAngels() != 3 {
		t.Errorf("expected 3 angels after prestige, got %d", e.GetAngels())
	}

	// Saldo dompet harus kembali ke 0
	if e.wallet.Balance() != 0.0 {
		t.Errorf("expected wallet balance to reset to 0, got %.2f", e.wallet.Balance())
	}

	// Upgrades harus ter-reset (IsPurchased = false)
	if e.upgrades[0].IsPurchased {
		t.Error("expected upgrade to be reset to unpurchased")
	}

	// Lifetime earnings tidak boleh direset
	if e.GetLifetimeEarnings() != 900.0 {
		t.Errorf("expected lifetime earnings to remain 900, got %.2f", e.GetLifetimeEarnings())
	}

	// 5. Cek apakah pengali investor (+5% per investor) aktif
	// Dengan 3 investor, pengali pendapatan global adalah 1.0 + 3 * 0.05 = 1.15
	// Lemonade Stand Level 1 BaseIncome = 1.0, pendapatan harus menjadi 1.15
	lemonade := e.businesses[0]
	if lemonade.Income() != 1.15 {
		t.Errorf("expected lemonade income to be 1.15 (+15%% investor bonus), got %.2f", lemonade.Income())
	}
}
