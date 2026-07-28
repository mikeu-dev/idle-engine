package engine

import (
	"testing"
	"time"
)

func TestTriggerSuperBoost(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Tambahkan saldo (saldo awal 4.0 + 100.0 = 104.0)
	e.wallet.Add(100.0)

	// Beli Lemonade Stand (biaya 4.0, sisa saldo = 100.0)
	e.BuyUpgrade(0)

	// Pastikan boostDuration awal adalah 0
	if e.GetBoostDuration() != 0 {
		t.Errorf("expected initial boost duration to be 0, got %v", e.GetBoostDuration())
	}

	// Pemicu Super Boost (biaya 50.0, sisa saldo = 50.0)
	success := e.TriggerSuperBoost()
	if !success {
		t.Fatal("expected Super Boost trigger to succeed")
	}

	// Verifikasi saldo terpotong (104 - 4 - 50 = 50.0)
	if e.GetWallet().Balance() != 50.0 {
		t.Errorf("expected wallet balance 50.0, got %f", e.GetWallet().Balance())
	}

	// Verifikasi sisa durasi boost adalah 30 detik
	if e.GetBoostDuration() != 30*time.Second {
		t.Errorf("expected boost duration to be 30s, got %v", e.GetBoostDuration())
	}

	// Uji berjalannya waktu di Update memotong durasi boost
	e.lastTick = time.Now()
	time.Sleep(10 * time.Millisecond)
	e.Update()

	if e.GetBoostDuration() >= 30*time.Second || e.GetBoostDuration() <= 0 {
		t.Errorf("expected boost duration to decrease, got %v", e.GetBoostDuration())
	}
}

func TestTriggerTimeWarp(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Tambahkan saldo (saldo awal 4.0 + 200.0 = 204.0)
	// Catatan: Menambahkan saldo > 100.0 otomatis membuka Achievement "pts_100" (+10% Pendapatan Global).
	e.wallet.Add(200.0)

	// Upgrade News secara manual lewat engine agar memotong saldo (biaya 20.0, sisa saldo = 184.0)
	successBuy := e.BuyUpgrade(1)
	if !successBuy {
		t.Fatal("expected to buy Newspaper upgrade successfully")
	}

	// Pemicu Time Warp (biaya 150.0, sisa saldo = 34.0)
	success := e.TriggerTimeWarp()
	if !success {
		t.Fatal("expected Time Warp to succeed")
	}

	// Kalkulasi ekspektasi saldo:
	// Saldo awal = 34.0 Poin
	// Pendapatan Newspaper (Lv 1) = 4.0 Poin * 1.1 (achievement bonus +10%) = 4.4 Poin
	// GPS Newspaper = 4.4 / 3.0s = 1.466667 Poin/detik
	// Hasil Time Warp 1 Jam = 1.466667 * 3600 = 5280.0 Poin
	// Saldo Akhir = 34.0 + 5280.0 = 5314.0 Poin
	expectedBalance := 5314.0
	diff := e.GetWallet().Balance() - expectedBalance
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("expected wallet balance to be around %f, got %f", expectedBalance, e.GetWallet().Balance())
	}
}
