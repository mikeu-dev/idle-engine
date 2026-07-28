package engine

import (
	"idle-engine/internal/domain/business"
	"idle-engine/internal/domain/economy"
	"sync"
	"time"
)

// Engine merepresentasikan orchestrator logika utama permainan.
type Engine struct {
	mu         sync.RWMutex
	wallet     *economy.Wallet
	businesses []*business.Business
	lastTick   time.Time
}

// NewEngine membuat instance Engine baru dengan setup bisnis awal.
func NewEngine() *Engine {
	// Buat wallet dengan modal awal 4 poin agar bisa beli Lemonade Stand segera
	wallet := economy.NewWallet(4.0)

	// Setup beberapa lini bisnis
	businesses := []*business.Business{
		// Bisnis 1: Murah, cepat, manual awalnya
		business.NewBusiness("lemonade", "Lemonade Stand", 4.0, 1.15, 1.0, 1*time.Second, false),
		// Bisnis 2: Sedang, otomatis
		business.NewBusiness("newspaper", "Newspaper Route", 20.0, 1.15, 4.0, 3*time.Second, true),
		// Bisnis 3: Mahal, lambat, otomatis, hasil besar
		business.NewBusiness("carwash", "Car Wash", 100.0, 1.15, 20.0, 6*time.Second, true),
	}

	return &Engine{
		wallet:     wallet,
		businesses: businesses,
		lastTick:   time.Now(),
	}
}

// Update memproses waktu yang berlalu dan memperbarui semua bisnis serta saldo wallet.
func (e *Engine) Update() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	delta := now.Sub(e.lastTick)
	e.lastTick = now

	// Update masing-masing bisnis dan tampung pendapatan
	totalRevenue := 0.0
	for _, b := range e.businesses {
		totalRevenue += b.Update(delta)
	}

	// Tambahkan pendapatan ke dompet
	if totalRevenue > 0 {
		e.wallet.Add(totalRevenue)
	}
}

// GetWallet mengembalikan wallet dari engine.
func (e *Engine) GetWallet() *economy.Wallet {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.wallet
}

// GetBusinesses mengembalikan semua bisnis yang terdaftar.
func (e *Engine) GetBusinesses() []*business.Business {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.businesses
}

// BuyUpgrade membeli atau menaikkan level bisnis tertentu jika saldo mencukupi.
func (e *Engine) BuyUpgrade(idx int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if idx < 0 || idx >= len(e.businesses) {
		return false
	}

	b := e.businesses[idx]
	cost := b.Cost()

	// Coba belanjakan uang dari wallet
	if e.wallet.Spend(cost) {
		b.Upgrade()
		return true
	}
	return false
}

// TriggerProduction memicu manual start produksi untuk bisnis non-otomatis.
func (e *Engine) TriggerProduction(idx int) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if idx < 0 || idx >= len(e.businesses) {
		return false
	}

	return e.businesses[idx].StartProduction()
}

