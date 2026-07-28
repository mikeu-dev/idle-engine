package engine

import (
	"idle-engine/internal/core/event"
	"idle-engine/internal/core/save"
	"idle-engine/internal/domain/business"
	"idle-engine/internal/domain/economy"
	"idle-engine/internal/domain/manager"
	"idle-engine/internal/domain/modifier"
	"idle-engine/internal/domain/upgrade"
	"sync"
	"time"
)

// Engine merepresentasikan orchestrator logika utama permainan.
type Engine struct {
	mu           sync.RWMutex
	wallet       *economy.Wallet
	businesses   []*business.Business
	upgrades     []*upgrade.Upgrade
	managers     []*manager.Manager
	achievements []*event.Achievement
	lastTick     time.Time
}

// NewEngine membuat instance Engine baru dengan setup bisnis awal, upgrade, manager, dan achievement.
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

	// Setup item upgrade bawaan
	upgrades := []*upgrade.Upgrade{
		upgrade.NewUpgrade("lemon_pitcher", "Lemon Pitcher", "Lemonade Stand 2x Pendapatan", 15.0, "lemonade", modifier.NewModifier(2.0, 1.0)),
		upgrade.NewUpgrade("newspaper_bag", "Newspaper Bag", "Newspaper Route 2x Kecepatan", 50.0, "newspaper", modifier.NewModifier(1.0, 2.0)),
		upgrade.NewUpgrade("power_washer", "Power Washer", "Car Wash 3x Pendapatan", 250.0, "carwash", modifier.NewModifier(3.0, 1.0)),
	}

	// Setup item manager otomatisasi bawaan (柠檬 stand manual, lainnya sudah otomatis bawaan)
	managers := []*manager.Manager{
		manager.NewManager("lemonade_mgr", "Lemonade Manager", "Mengotomatiskan Lemonade Stand secara permanen", 100.0, "lemonade"),
	}

	// Setup pencapaian (Achievements) bawaan
	achievements := []*event.Achievement{
		event.NewAchievement("pts_100", "Poin Pemula", "Kumpulkan 100 Poin (Bonus +10% pendapatan global)", "balance", "", 100.0, 0.10),
		event.NewAchievement("lemon_10", "Lemonade Tycoon", "Lemonade Stand Level 10 (Bonus +20% pendapatan Lemonade)", "level", "lemonade", 10.0, 0.20),
		event.NewAchievement("news_10", "Newspaper Tycoon", "Newspaper Route Level 10 (Bonus +20% pendapatan Newspaper)", "level", "newspaper", 10.0, 0.20),
	}

	return &Engine{
		wallet:       wallet,
		businesses:   businesses,
		upgrades:     upgrades,
		managers:     managers,
		achievements: achievements,
		lastTick:     time.Now(),
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

	// Periksa pencapaian baru setelah saldo berubah
	e.checkAchievements()
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

// GetUpgrades mengembalikan daftar semua upgrade.
func (e *Engine) GetUpgrades() []*upgrade.Upgrade {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.upgrades
}

// GetManagers mengembalikan daftar semua manager.
func (e *Engine) GetManagers() []*manager.Manager {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.managers
}

// GetAchievements mengembalikan daftar semua pencapaian.
func (e *Engine) GetAchievements() []*event.Achievement {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.achievements
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
		e.checkAchievements() // Periksa pencapaian setelah level bisnis berubah
		return true
	}
	return false
}

// BuyUpgradeCard membeli item upgrade tertentu berdasarkan indeks jika saldo mencukupi.
func (e *Engine) BuyUpgradeCard(idx int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if idx < 0 || idx >= len(e.upgrades) {
		return false
	}

	upg := e.upgrades[idx]
	if upg.IsPurchased {
		return false
	}

	if e.wallet.Spend(upg.Cost) {
		upg.Purchase()
		e.applyModifiers() // Hitung ulang modifiers bisnis setelah ada upgrade baru
		e.checkAchievements() // Periksa pencapaian
		return true
	}
	return false
}

// BuyManager mempekerjakan manager tertentu berdasarkan indeks jika saldo mencukupi.
func (e *Engine) BuyManager(idx int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if idx < 0 || idx >= len(e.managers) {
		return false
	}

	m := e.managers[idx]
	if m.IsHired {
		return false
	}

	if e.wallet.Spend(m.Cost) {
		m.Hire()
		e.applyAutomation()
		e.checkAchievements()
		return true
	}
	return false
}

// applyModifiers menghitung ulang dan menerapkan modifier dari upgrade dan achievements yang diperoleh.
// Catatan: Pemanggil harus menahan lock mu.
func (e *Engine) applyModifiers() {
	revMults := make(map[string]float64)
	speedMults := make(map[string]float64)

	for _, b := range e.businesses {
		revMults[b.ID] = 1.0
		speedMults[b.ID] = 1.0
	}

	// 1. Terapkan upgrade aktif
	for _, upg := range e.upgrades {
		if upg.IsPurchased {
			revMults[upg.TargetBusinessID] *= upg.Effect.RevenueMultiplier
			speedMults[upg.TargetBusinessID] *= upg.Effect.SpeedMultiplier
		}
	}

	// 2. Terapkan bonus pencapaian (Achievements) aktif
	for _, ach := range e.achievements {
		if ach.IsUnlocked {
			if ach.TargetBusinessID == "" {
				// Bonus global
				for _, b := range e.businesses {
					revMults[b.ID] *= (1.0 + ach.BonusMultiplier)
				}
			} else {
				// Bonus spesifik lini bisnis
				revMults[ach.TargetBusinessID] *= (1.0 + ach.BonusMultiplier)
			}
		}
	}

	// Terapkan ke masing-masing bisnis
	for _, b := range e.businesses {
		b.SetModifiers(revMults[b.ID], speedMults[b.ID])
	}
}

// applyAutomation menerapkan otomatisasi berdasarkan manager yang disewa.
// Catatan: Pemanggil harus menahan lock mu.
func (e *Engine) applyAutomation() {
	for _, m := range e.managers {
		if m.IsHired {
			for _, b := range e.businesses {
				if b.ID == m.TargetBusinessID {
					b.SetAutomated(true)
				}
			}
		}
	}
}

// checkAchievements memeriksa kondisi pencapaian dan mengaktifkan bonus jika terpenuhi.
// Catatan: Pemanggil harus menahan lock mu.
func (e *Engine) checkAchievements() {
	businessLevels := make(map[string]int)
	for _, b := range e.businesses {
		businessLevels[b.ID] = b.GetLevel()
	}

	balance := e.wallet.Balance()
	unlockedAny := false

	for _, ach := range e.achievements {
		if ach.CheckCondition(balance, businessLevels) {
			ach.Unlock()
			unlockedAny = true
		}
	}

	if unlockedAny {
		e.applyModifiers() // Hitung ulang modifiers karena ada bonus pencapaian baru
	}
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

	var purchasedUpgrades []string
	for _, upg := range e.upgrades {
		if upg.IsPurchased {
			purchasedUpgrades = append(purchasedUpgrades, upg.ID)
		}
	}

	var hiredManagers []string
	for _, m := range e.managers {
		if m.IsHired {
			hiredManagers = append(hiredManagers, m.ID)
		}
	}

	var unlockedAchievements []string
	for _, ach := range e.achievements {
		if ach.IsUnlocked {
			unlockedAchievements = append(unlockedAchievements, ach.ID)
		}
	}

	return &save.SaveState{
		Balance:      e.wallet.Balance(),
		Businesses:   bizStates,
		Upgrades:     purchasedUpgrades,
		Managers:     hiredManagers,
		Achievements: unlockedAchievements,
		Timestamp:    time.Now(),
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

	// 3. Pulihkan state upgrade
	purchasedUpgMap := make(map[string]bool)
	for _, id := range state.Upgrades {
		purchasedUpgMap[id] = true
	}
	for _, upg := range e.upgrades {
		upg.IsPurchased = purchasedUpgMap[upg.ID]
	}

	// 4. Pulihkan state manager
	hiredMgrMap := make(map[string]bool)
	for _, id := range state.Managers {
		hiredMgrMap[id] = true
	}
	for _, m := range e.managers {
		m.IsHired = hiredMgrMap[m.ID]
	}
	e.applyAutomation()

	// 5. Pulihkan state achievement
	unlockedAchMap := make(map[string]bool)
	for _, id := range state.Achievements {
		unlockedAchMap[id] = true
	}
	for _, ach := range e.achievements {
		ach.IsUnlocked = unlockedAchMap[ach.ID]
	}

	// 6. Hitung ulang modifiers
	e.applyModifiers()

	// 7. Hitung pendapatan offline (maksimal 12 jam)
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




