package engine

import (
	"testing"
	"time"
)

func TestOfflineProgress(t *testing.T) {
	// 1. Inisialisasi Engine baru
	eng := NewEngine()

	// Lemonade Stand = indeks 0, Newspaper Route = indeks 1
	// Beli level 1 untuk Newspaper Route (otomatis, biaya 20, saldo awal 4)
	// Kita tambahkan saldo agar bisa membeli
	eng.GetWallet().Add(100.0)
	
	success := eng.BuyUpgrade(1) // Beli Newspaper Route (Lv 1)
	if !success {
		t.Fatal("failed to buy Newspaper Route upgrade")
	}

	// Sisa saldo setelah beli: 104 - 20 = 84
	initialBalance := eng.GetWallet().Balance()
	if initialBalance != 84.0 {
		t.Errorf("expected balance 84.0, got %f", initialBalance)
	}

	// 2. Ekspor state saat ini
	state := eng.ExportState()

	// Manipulasi stempel waktu 1 jam ke masa lalu (3600 detik)
	// Newspaper Route durasi = 3 detik, BaseIncome = 4.0.
	// Dalam 1 jam: 3600 / 3 = 1200 siklus.
	// Total pendapatan offline: 1200 * 4.0 = 4800.0 poin.
	state.Timestamp = state.Timestamp.Add(-1 * time.Hour)

	// Buat engine baru untuk mensimulasikan startup game
	engNew := NewEngine()

	// 3. Impor state yang dimanipulasi
	offlineRevenue := engNew.ImportState(state)

	expectedRevenue := 4800.0
	if offlineRevenue != expectedRevenue {
		t.Errorf("expected offline revenue %f, got %f", expectedRevenue, offlineRevenue)
	}

	// Saldo akhir harus: saldo yang disimpan (84.0) + pendapatan offline (4800.0) = 4884.0
	expectedTotalBalance := 84.0 + 4800.0
	actualBalance := engNew.GetWallet().Balance()
	if actualBalance != expectedTotalBalance {
		t.Errorf("expected total balance %f, got %f", expectedTotalBalance, actualBalance)
	}

	// Pastikan level Newspaper Route terpulihkan ke level 1
	if level := engNew.GetBusinesses()[1].GetLevel(); level != 1 {
		t.Errorf("expected Newspaper Route level 1, got %d", level)
	}
}

func TestOfflineProgressCap(t *testing.T) {
	// Menguji batas maksimal pendapatan offline yaitu 12 jam.
	eng := NewEngine()
	eng.GetWallet().Add(100.0)
	eng.BuyUpgrade(1) // Beli Newspaper Route (Lv 1)

	state := eng.ExportState()

	// Manipulasi stempel waktu 24 jam ke masa lalu (melebihi limit 12 jam)
	// Kita batasi maksimal 12 jam (43200 detik).
	// Dalam 12 jam: 43200 / 3 = 14400 siklus.
	// Total pendapatan offline max: 14400 * 4.0 = 57600.0 poin.
	state.Timestamp = state.Timestamp.Add(-24 * time.Hour)

	engNew := NewEngine()
	offlineRevenue := engNew.ImportState(state)

	expectedRevenue := 57600.0 // Terbatas 12 jam
	if offlineRevenue != expectedRevenue {
		t.Errorf("expected offline revenue capped at 12 hours (%f), got %f", expectedRevenue, offlineRevenue)
	}
}
