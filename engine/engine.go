package engine

import (
	"idle-engine/internal/core/event"
	"idle-engine/internal/core/save"
	gametime "idle-engine/internal/core/time"
	"idle-engine/internal/domain/business"
	"idle-engine/internal/domain/economy"
	"idle-engine/internal/domain/manager"
	"idle-engine/internal/domain/modifier"
	"idle-engine/internal/domain/upgrade"
	"idle-engine/internal/infrastructure/config"
	"idle-engine/pkg/mathutil"
	"math"
	"sync"
	"time"
)

// Engine merepresentasikan orchestrator logika utama permainan.
type Engine struct {
	mu               sync.RWMutex
	wallet           *economy.Wallet
	businesses       []*business.Business
	upgrades         []*upgrade.Upgrade
	managers         []*manager.Manager
	achievements     []*event.Achievement
	lifetimeEarnings float64
	angels           int
	lastTick         time.Time
}

// NewEngineWithConfig membuat instance Engine baru berdasarkan berkas konfigurasi YAML di cfgPath.
// Jika file tidak dapat dibaca atau di-decode, engine akan menggunakan konfigurasi fallback bawaan agar unit test tetap kompatibel.
func NewEngineWithConfig(cfgPath string) *Engine {
	// Buat wallet dengan modal awal 4 poin agar bisa beli Lemonade Stand segera
	wallet := economy.NewWallet(4.0)

	var businesses []*business.Business
	var upgrades []*upgrade.Upgrade
	var managers []*manager.Manager
	var achievements []*event.Achievement

	// Coba muat konfigurasi dari file
	cfg, err := config.LoadConfig(cfgPath)
	if err == nil && len(cfg.Businesses) > 0 {
		// Konfigurasi dinamis dari YAML
		for _, bc := range cfg.Businesses {
			duration := time.Duration(bc.DurationMs) * time.Millisecond
			b := business.NewBusiness(bc.ID, bc.Name, bc.BaseCost, bc.CostMultiplier, bc.BaseIncome, duration, bc.IsAutomated)
			businesses = append(businesses, b)
		}

		for _, uc := range cfg.Upgrades {
			upgrades = append(upgrades, upgrade.NewUpgrade(
				uc.ID,
				uc.Name,
				uc.Description,
				uc.Cost,
				uc.TargetBusinessID,
				modifier.NewModifier(uc.RevenueMultiplier, uc.SpeedMultiplier),
			))
		}

		for _, mc := range cfg.Managers {
			managers = append(managers, manager.NewManager(
				mc.ID,
				mc.Name,
				mc.Description,
				mc.Cost,
				mc.TargetBusinessID,
			))
		}

		for _, ac := range cfg.Achievements {
			achievements = append(achievements, event.NewAchievement(
				ac.ID,
				ac.Name,
				ac.Description,
				ac.ConditionType,
				ac.TargetBusinessID,
				ac.TargetValue,
				ac.BonusMultiplier,
			))
		}
	} else {
		// Fallback bawaan (hardcoded) demi kompatibilitas mundur
		businesses = []*business.Business{
			business.NewBusiness("lemonade", "Lemonade Stand", 4.0, 1.15, 1.0, 1*time.Second, false),
			business.NewBusiness("newspaper", "Newspaper Route", 20.0, 1.15, 4.0, 3*time.Second, true),
			business.NewBusiness("carwash", "Car Wash", 100.0, 1.15, 20.0, 6*time.Second, true),
		}

		upgrades = []*upgrade.Upgrade{
			upgrade.NewUpgrade("lemon_pitcher", "Lemon Pitcher", "Lemonade Stand 2x Pendapatan", 15.0, "lemonade", modifier.NewModifier(2.0, 1.0)),
			upgrade.NewUpgrade("newspaper_bag", "Newspaper Bag", "Newspaper Route 2x Kecepatan", 50.0, "newspaper", modifier.NewModifier(1.0, 2.0)),
			upgrade.NewUpgrade("power_washer", "Power Washer", "Car Wash 3x Pendapatan", 250.0, "carwash", modifier.NewModifier(3.0, 1.0)),
		}

		managers = []*manager.Manager{
			manager.NewManager("lemonade_mgr", "Lemonade Manager", "Mengotomatiskan Lemonade Stand secara permanen", 100.0, "lemonade"),
		}

		achievements = []*event.Achievement{
			event.NewAchievement("pts_100", "Poin Pemula", "Kumpulkan 100 Poin (Bonus +10% pendapatan global)", "balance", "", 100.0, 0.10),
			event.NewAchievement("lemon_10", "Lemonade Tycoon", "Lemonade Stand Level 10 (Bonus +20% pendapatan Lemonade)", "level", "lemonade", 10.0, 0.20),
			event.NewAchievement("news_10", "Newspaper Tycoon", "Newspaper Route Level 10 (Bonus +20% pendapatan Newspaper)", "level", "newspaper", 10.0, 0.20),
		}
	}

	return &Engine{
		wallet:           wallet,
		businesses:       businesses,
		upgrades:         upgrades,
		managers:         managers,
		achievements:     achievements,
		lifetimeEarnings: 4.0,
		angels:           0,
		lastTick:         time.Now(),
	}
}

// NewEngine membuat instance Engine baru dengan konfigurasi bawaan game_config.yaml.
func NewEngine() *Engine {
	return NewEngineWithConfig("internal/infrastructure/config/game_config.yaml")
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

	// Tambahkan pendapatan ke dompet dan lifetime earnings
	if totalRevenue > 0 {
		e.wallet.Add(totalRevenue)
		e.lifetimeEarnings += totalRevenue
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

// GetLifetimeEarnings mengembalikan akumulasi pendapatan sepanjang masa.
func (e *Engine) GetLifetimeEarnings() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lifetimeEarnings
}

// GetAngels mengembalikan jumlah Angel Investors saat ini.
func (e *Engine) GetAngels() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.angels
}

// CalculateAngelsToClaim menghitung berapa banyak investor yang bisa diperoleh jika mereset progres sekarang.
func (e *Engine) CalculateAngelsToClaim() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Rumus uji coba sandbox: floor(sqrt(LifetimeEarnings / 100)) - Angels
	if e.lifetimeEarnings < 100.0 {
		return 0
	}
	totalAngels := int(math.Floor(math.Sqrt(e.lifetimeEarnings / 100.0)))
	claimable := totalAngels - e.angels
	if claimable < 0 {
		return 0
	}
	return claimable
}

// ClaimPrestige melakukan reset progres permainan dan mengklaim Angel Investors baru.
func (e *Engine) ClaimPrestige() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.lifetimeEarnings < 100.0 {
		return false
	}
	totalAngels := int(math.Floor(math.Sqrt(e.lifetimeEarnings / 100.0)))
	claimable := totalAngels - e.angels
	if claimable <= 0 {
		return false
	}

	// 1. Klaim investor baru
	e.angels += claimable

	// 2. Reset Saldo Wallet
	e.wallet = economy.NewWallet(0)

	// 3. Reset Level Bisnis ke tingkat awal
	for _, b := range e.businesses {
		if b.ID == "lemonade" {
			b.LoadState(1, false, 0)
			b.SetAutomated(false)
		} else {
			b.LoadState(0, false, 0)
			if b.ID == "newspaper" || b.ID == "carwash" {
				b.SetAutomated(true) // Newspaper dan Car Wash otomatis bawaan saat level > 0
			}
		}
	}

	// 4. Reset status pembelian Upgrade Card
	for _, upg := range e.upgrades {
		upg.IsPurchased = false
	}

	// 5. Reset status perekrutan Manager
	for _, m := range e.managers {
		m.IsHired = false
	}

	// Catatan: Pencapaian (Achievements) TIDAK DIRESET agar bertahan permanen.

	// 6. Hitung ulang modifiers (akan mengintegrasikan bonus pengali Angel Investors)
	e.applyModifiers()

	// Reset waktu update
	e.lastTick = time.Now()

	return true
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

// BuyUpgradeMax membeli level bisnis sebanyak mungkin (Buy Max) berdasarkan saldo yang tersedia.
func (e *Engine) BuyUpgradeMax(idx int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if idx < 0 || idx >= len(e.businesses) {
		return false
	}

	b := e.businesses[idx]
	balance := e.wallet.Balance()

	levels, cost := mathutil.CalculateMaxLevelsAffordable(b.BaseCost, b.CostMultiplier, b.GetLevel(), balance)
	if levels <= 0 {
		return false
	}

	if e.wallet.Spend(cost) {
		b.UpgradeMany(levels)
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

	// 3. Terapkan bonus Angel Investors (+5% pendapatan global per investor)
	if e.angels > 0 {
		angelMultiplier := 1.0 + float64(e.angels)*0.05
		for _, b := range e.businesses {
			revMults[b.ID] *= angelMultiplier
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
		Balance:          e.wallet.Balance(),
		Businesses:       bizStates,
		Upgrades:         purchasedUpgrades,
		Managers:         hiredManagers,
		Achievements:     unlockedAchievements,
		LifetimeEarnings: e.lifetimeEarnings,
		Angels:           e.angels,
		Timestamp:        time.Now(),
	}
}

// ImportState memuat state penyimpanan ke dalam engine dan menghitung pendapatan offline dengan proteksi NTP anti-cheat.
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

	// 6. Pulihkan total pendapatan dan investor (backward compatible)
	e.lifetimeEarnings = state.LifetimeEarnings
	if e.lifetimeEarnings < e.wallet.Balance() {
		e.lifetimeEarnings = e.wallet.Balance()
	}
	e.angels = state.Angels

	// 7. Hitung ulang modifiers
	e.applyModifiers()

	// 8. Hitung pendapatan offline (maksimal 12 jam) dengan proteksi anti-cheat
	now := time.Now()
	offlineDuration := now.Sub(state.Timestamp)

	// Proteksi 1: Gunakan NTP jika terhubung
	ntpTime, ntpErr := gametime.GetNetworkTime("pool.ntp.org", 1500*time.Millisecond)
	if ntpErr == nil {
		// Validasi apakah waktu lokal akurat (toleransi 5 menit)
		isLocalValid := gametime.IsSystemTimeValid(now, ntpTime, 5*time.Minute)
		if !isLocalValid {
			// Jam lokal diubah secara tidak akurat/manipulatif, paksa gunakan selisih NTP tepercaya!
			offlineDuration = ntpTime.Sub(state.Timestamp)
		}

		// Validasi apakah waktu NTP berada sebelum waktu simpan terakhir (cheat jam dimundurkan)
		if ntpTime.Before(state.Timestamp) {
			offlineDuration = 0
		}
	} else {
		// Offline fallback ke waktu lokal
		// Proteksi 2: Deteksi jika waktu lokal dimundurkan ke belakang waktu simpan terakhir
		if now.Before(state.Timestamp) {
			offlineDuration = 0
		}
	}

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

// GetTotalGPS mengembalikan total pendapatan per detik dari seluruh bisnis otomatis yang dimiliki pemain.
func (e *Engine) GetTotalGPS() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	total := 0.0
	for _, b := range e.businesses {
		if b.IsOwned() && b.IsAutomated {
			total += b.GetGPS()
		}
	}
	return total
}





