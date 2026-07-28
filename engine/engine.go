package engine

import (
	"idle-engine/internal/core/save"
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

// ExportState mengekspor state engine saat ini ke dalam bentuk SaveState.
func (e *Engine) ExportState() *save.SaveState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	bizStates := make([]save.BizState, len(e.businesses))
	for i, b := range e.businesses {
		bizStates[i] = save.BizState{
			ID:         b.ID,
			Level:      b.GetLevel(),
			IsActive:   b.GetIsActive(),
			ProgressNs: b.GetProgress(),
		}
	}

	return &save.SaveState{
		Balance:    e.wallet.Balance(),
		Businesses: bizStates,
		Timestamp:  time.Now(),
	}
}

// ImportState memuat state penyimpanan ke dalam engine dan menghitung pendapatan offline.
// Mengembalikan total pendapatan offline yang berhasil dikumpulkan (maksimal 12 jam offline).
func (e *Engine) ImportState(state *save.SaveState) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Pulihkan saldo wallet
	e.wallet = economy.NewWallet(state.Balance)

	// 2. Pulihkan state bisnis
	bizStateMap := make(map[string]save.BizState)
	for _, bs := range state.Businesses {
		bizStateMap[bs.ID] = bs
	}

	for _, b := range e.businesses {
		if bs, ok := bizStateMap[b.ID]; ok {
			b.LoadState(bs.Level, bs.IsActive, bs.ProgressNs)
		}
	}

	// 3. Hitung pendapatan offline (maksimal 12 jam)
	now := time.Now()
	offlineDuration := now.Sub(state.Timestamp)

	maxOffline := 12 * time.Hour
	if offlineDuration > maxOffline {
		offlineDuration = maxOffline
	}

	offlineRevenue := 0.0
	if offlineDuration > 0 {
		for _, b := range e.businesses {
			offlineRevenue += b.Update(offlineDuration)
		}
	}

	if offlineRevenue > 0 {
		e.wallet.Add(offlineRevenue)
	}

	// Reset lastTick ke waktu sekarang agar tidak mendobel durasi offline pada update frame berikutnya
	e.lastTick = now

	return offlineRevenue
}


