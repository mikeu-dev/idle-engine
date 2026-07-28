package business

import (
	"testing"
	"time"
)

func TestBusinessManual(t *testing.T) {
	// Buat bisnis manual dengan durasi 1 detik
	b := NewBusiness("lemonade", "Lemonade Stand", 10.0, 1.15, 2.0, time.Second, false)

	if b.IsOwned() {
		t.Error("expected business to be unowned initially")
	}
	if b.Cost() != 10.0 {
		t.Errorf("expected cost 10.0, got %f", b.Cost())
	}

	// Upgrade ke level 1
	b.Upgrade()
	if !b.IsOwned() || b.GetLevel() != 1 {
		t.Errorf("expected level 1, got %d", b.GetLevel())
	}
	if b.GetIsActive() {
		t.Error("expected manual business to be inactive initially after upgrade")
	}

	// Jalankan produksi manual
	started := b.StartProduction()
	if !started || !b.GetIsActive() {
		t.Error("failed to start production")
	}

	// Update dengan setengah durasi (500ms)
	revenue := b.Update(500 * time.Millisecond)
	if revenue != 0.0 {
		t.Errorf("expected 0 revenue, got %f", revenue)
	}
	if pct := b.GetProgressPercent(); pct != 0.5 {
		t.Errorf("expected progress 0.5, got %f", pct)
	}

	// Update dengan sisa durasi (600ms) untuk menyelesaikan siklus
	revenue = b.Update(600 * time.Millisecond)
	if revenue != 2.0 {
		t.Errorf("expected revenue 2.0, got %f", revenue)
	}
	if b.GetIsActive() {
		t.Error("expected manual business to deactivate after finishing production")
	}
}

func TestBusinessAutomated(t *testing.T) {
	// Buat bisnis otomatis dengan durasi 2 detik
	b := NewBusiness("newspaper", "Newspaper Delivery", 100.0, 1.15, 10.0, 2*time.Second, true)

	// Upgrade ke level 1
	b.Upgrade()
	if !b.GetIsActive() {
		t.Error("expected automated business to start active upon upgrade")
	}

	// Update dengan durasi panjang (5 detik) - harus menyelesaikan 2 siklus (4 detik) dan menyisakan progress 1 detik
	revenue := b.Update(5 * time.Second)
	expectedRevenue := 20.0 // 10.0 * level 1 * 2 cycles
	if revenue != expectedRevenue {
		t.Errorf("expected revenue %f, got %f", expectedRevenue, revenue)
	}
	if pct := b.GetProgressPercent(); pct != 0.5 { // 1 detik dari 2 detik durasi = 0.5
		t.Errorf("expected progress 0.5, got %f", pct)
	}
}
