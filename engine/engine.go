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

// RandomEvent represents a short-duration economic random event.
type RandomEvent struct {
	ID          string
	Title       string
	Description string
	TargetBizID string // If empty (""), it is global
	EffectType  string // "revenue" or "speed"
	Multiplier  float64
}

// Engine represents the main game logic orchestrator.
type Engine struct {
	mu                      sync.RWMutex
	wallet                  *economy.Wallet
	businesses              []*business.Business
	upgrades                []*upgrade.Upgrade
	managers                []*manager.Manager
	achievements            []*event.Achievement
	lifetimeEarnings        float64
	angels                  int
	boostDuration           time.Duration
	activeEvent             *RandomEvent
	eventDuration           time.Duration
	timeSinceLastEventCheck time.Duration
	lastTick                time.Time
}

// NewEngineWithConfig creates a new Engine instance based on the YAML config file at cfgPath.
// If the file cannot be loaded or decoded, the engine uses fallback configuration for unit test compatibility.
func NewEngineWithConfig(cfgPath string) *Engine {
	// Create wallet with initial 4 points so players can buy the Lemonade Stand immediately
	wallet := economy.NewWallet(4.0)

	var businesses []*business.Business
	var upgrades []*upgrade.Upgrade
	var managers []*manager.Manager
	var achievements []*event.Achievement

	// Try to load configuration from file
	cfg, err := config.LoadConfig(cfgPath)
	if err == nil && len(cfg.Businesses) > 0 {
		// Dynamic configuration from YAML
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
		// Default fallback (hardcoded) for backward compatibility
		businesses = []*business.Business{
			business.NewBusiness("lemonade", "Lemonade Stand", 4.0, 1.15, 1.0, 1*time.Second, false),
			business.NewBusiness("newspaper", "Newspaper Route", 20.0, 1.15, 4.0, 3*time.Second, true),
			business.NewBusiness("carwash", "Car Wash", 100.0, 1.15, 20.0, 6*time.Second, true),
		}

		upgrades = []*upgrade.Upgrade{
			upgrade.NewUpgrade("lemon_pitcher", "Lemon Pitcher", "Lemonade Stand 2x Revenue", 15.0, "lemonade", modifier.NewModifier(2.0, 1.0)),
			upgrade.NewUpgrade("newspaper_bag", "Newspaper Bag", "Newspaper Route 2x Speed", 50.0, "newspaper", modifier.NewModifier(1.0, 2.0)),
			upgrade.NewUpgrade("power_washer", "Power Washer", "Car Wash 3x Revenue", 250.0, "carwash", modifier.NewModifier(3.0, 1.0)),
		}

		managers = []*manager.Manager{
			manager.NewManager("lemonade_mgr", "Lemonade Manager", "Automates Lemonade Stand production permanently", 100.0, "lemonade"),
		}

		achievements = []*event.Achievement{
			event.NewAchievement("pts_100", "Beginner Points", "Collect 100 Points (Bonus +10% global revenue)", "balance", "", 100.0, 0.10),
			event.NewAchievement("lemon_10", "Lemonade Tycoon", "Lemonade Stand Level 10 (Bonus +20% Lemonade revenue)", "level", "lemonade", 10.0, 0.20),
			event.NewAchievement("news_10", "Newspaper Tycoon", "Newspaper Route Level 10 (Bonus +20% Newspaper revenue)", "level", "newspaper", 10.0, 0.20),
		}
	}

	return &Engine{
		wallet:                  wallet,
		businesses:              businesses,
		upgrades:                upgrades,
		managers:                managers,
		achievements:            achievements,
		lifetimeEarnings:        4.0,
		angels:                  0,
		boostDuration:           0,
		activeEvent:             nil,
		eventDuration:           0,
		timeSinceLastEventCheck: 0,
		lastTick:                time.Now(),
	}
}

// NewEngine creates a new Engine instance with the default game_config.yaml configuration.
func NewEngine() *Engine {
	return NewEngineWithConfig("internal/infrastructure/config/game_config.yaml")
}

// Update processes the elapsed time and updates all businesses and wallet balance.
func (e *Engine) Update() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	delta := now.Sub(e.lastTick)
	e.lastTick = now

	// Update remaining Super Boost duration
	if e.boostDuration > 0 {
		e.boostDuration -= delta
		if e.boostDuration <= 0 {
			e.boostDuration = 0
			e.applyModifiers() // Recalculate modifiers because boost ended
		}
	}

	// Update remaining Random Event duration if any
	if e.activeEvent != nil {
		e.eventDuration -= delta
		if e.eventDuration <= 0 {
			e.activeEvent = nil
			e.eventDuration = 0
			e.applyModifiers() // Restore normal multipliers
		}
	} else {
		e.timeSinceLastEventCheck += delta
		if e.timeSinceLastEventCheck >= 45*time.Second {
			e.timeSinceLastEventCheck = 0
			if time.Now().UnixNano()%2 == 0 {
				e.triggerRandomEvent()
			}
		}
	}

	// Update each business and collect revenue
	totalRevenue := 0.0
	for _, b := range e.businesses {
		totalRevenue += b.Update(delta)
	}

	// Add revenue to wallet and lifetime earnings
	if totalRevenue > 0 {
		e.wallet.Add(totalRevenue)
		e.lifetimeEarnings += totalRevenue
	}

	// Check new achievements after balance changes
	e.checkAchievements()
}

// GetWallet returns the engine's wallet.
func (e *Engine) GetWallet() *economy.Wallet {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.wallet
}

// GetBusinesses returns all registered businesses.
func (e *Engine) GetBusinesses() []*business.Business {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.businesses
}

// GetUpgrades returns all upgrade items.
func (e *Engine) GetUpgrades() []*upgrade.Upgrade {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.upgrades
}

// GetManagers returns all managers.
func (e *Engine) GetManagers() []*manager.Manager {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.managers
}

// GetAchievements returns all achievements.
func (e *Engine) GetAchievements() []*event.Achievement {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.achievements
}

// GetLifetimeEarnings returns the cumulative lifetime earnings.
func (e *Engine) GetLifetimeEarnings() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lifetimeEarnings
}

// GetAngels returns the current number of Angel Investors.
func (e *Engine) GetAngels() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.angels
}

// GetBoostDuration returns the remaining Super Boost duration in a thread-safe manner.
func (e *Engine) GetBoostDuration() time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.boostDuration
}

// TriggerSuperBoost activates the Super Boost for 30s by spending the cost from the wallet.
func (e *Engine) TriggerSuperBoost() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	cost := 50.0
	if e.wallet.Spend(cost) {
		e.boostDuration += 30 * time.Second
		e.applyModifiers() // Recalculate business speeds with booster
		return true
	}
	return false
}

// TriggerTimeWarp grants 1 hour of instant passive automated income by spending the cost from the wallet.
func (e *Engine) TriggerTimeWarp() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	cost := 150.0
	if e.wallet.Spend(cost) {
		// Calculate total GPS of all automated businesses owned by the player
		totalGPS := 0.0
		for _, b := range e.businesses {
			if b.IsOwned() && b.IsAutomated {
				totalGPS += b.GetGPS()
			}
		}

		// Instant 1 hour of revenue (3600 seconds)
		instantRevenue := totalGPS * 3600.0
		if instantRevenue > 0 {
			e.wallet.Add(instantRevenue)
			e.lifetimeEarnings += instantRevenue
			e.checkAchievements()
		}
		return true
	}
	return false
}

// GetActiveEvent returns the currently active random event and its remaining duration in a thread-safe manner.
func (e *Engine) GetActiveEvent() (*RandomEvent, time.Duration) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activeEvent, e.eventDuration
}

// TriggerEventByID triggers a specific random event instantly by its ID for unit test purposes.
func (e *Engine) TriggerEventByID(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	events := map[string]*RandomEvent{
		"summer": {
			ID:          "summer",
			Title:       "Hot Summer",
			Description: "Lemonade Stand revenue 3x!",
			TargetBizID: "lemonade",
			EffectType:  "revenue",
			Multiplier:  3.0,
		},
		"paper_day": {
			ID:          "paper_day",
			Title:       "Paperboy Day",
			Description: "Newspaper Route speed 2x!",
			TargetBizID: "newspaper",
			EffectType:  "speed",
			Multiplier:  2.0,
		},
		"rain_storm": {
			ID:          "rain_storm",
			Title:       "Muddy Rainstorm",
			Description: "Car Wash revenue 4x!",
			TargetBizID: "carwash",
			EffectType:  "revenue",
			Multiplier:  4.0,
		},
		"power_outage": {
			ID:          "power_outage",
			Title:       "City Power Outage",
			Description: "All business speeds reduced to 0.5x!",
			TargetBizID: "",
			EffectType:  "speed",
			Multiplier:  0.5,
		},
	}

	evt, exists := events[id]
	if !exists {
		return false
	}

	e.activeEvent = evt
	e.eventDuration = 30 * time.Second
	e.applyModifiers()
	return true
}

// triggerRandomEvent triggers one of the random events for 30s.
// Note: Caller must hold Lock on mu.
func (e *Engine) triggerRandomEvent() {
	events := []*RandomEvent{
		{
			ID:          "summer",
			Title:       "Hot Summer",
			Description: "Lemonade Stand revenue 3x!",
			TargetBizID: "lemonade",
			EffectType:  "revenue",
			Multiplier:  3.0,
		},
		{
			ID:          "paper_day",
			Title:       "Paperboy Day",
			Description: "Newspaper Route speed 2x!",
			TargetBizID: "newspaper",
			EffectType:  "speed",
			Multiplier:  2.0,
		},
		{
			ID:          "rain_storm",
			Title:       "Muddy Rainstorm",
			Description: "Car Wash revenue 4x!",
			TargetBizID: "carwash",
			EffectType:  "revenue",
			Multiplier:  4.0,
		},
		{
			ID:          "power_outage",
			Title:       "City Power Outage",
			Description: "All business speeds reduced to 0.5x!",
			TargetBizID: "",
			EffectType:  "speed",
			Multiplier:  0.5,
		},
	}

	idx := int(time.Now().UnixNano() % int64(len(events)))
	e.activeEvent = events[idx]
	e.eventDuration = 30 * time.Second
	e.applyModifiers()
}

// CalculateAngelsToClaim calculates how many angels can be claimed if prestige reset is performed now.
func (e *Engine) CalculateAngelsToClaim() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Sandbox test formula: floor(sqrt(LifetimeEarnings / 100)) - Angels
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

// ClaimPrestige resets game progress and claims new Angel Investors.
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

	// 1. Claim new investors
	e.angels += claimable

	// 2. Reset wallet balance
	e.wallet = economy.NewWallet(0)

	// 3. Reset business levels to initial levels
	for _, b := range e.businesses {
		if b.ID == "lemonade" {
			b.LoadState(1, false, 0)
			b.SetAutomated(false)
		} else {
			b.LoadState(0, false, 0)
			if b.ID == "newspaper" || b.ID == "carwash" {
				b.SetAutomated(true) // Newspaper and Car Wash are automated by default once level > 0
			}
		}
	}

	// 4. Reset upgrade cards purchase status
	for _, upg := range e.upgrades {
		upg.IsPurchased = false
	}

	// 5. Reset hired managers
	for _, m := range e.managers {
		m.IsHired = false
	}

	// Note: Achievements are NOT reset and persist permanently.

	// 6. Recalculate modifiers (integrating Angel Investors global multiplier bonus)
	e.applyModifiers()

	// Reset update timestamp
	e.lastTick = time.Now()

	return true
}

// BuyUpgrade purchases or upgrades a specific business if balance is sufficient.
func (e *Engine) BuyUpgrade(idx int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if idx < 0 || idx >= len(e.businesses) {
		return false
	}

	b := e.businesses[idx]
	cost := b.Cost()

	// Try to spend from wallet
	if e.wallet.Spend(cost) {
		b.Upgrade()
		e.checkAchievements() // Check achievements after business level change
		return true
	}
	return false
}

// BuyUpgradeMax purchases as many levels of a business as possible (Buy Max) based on the available balance.
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
		e.checkAchievements() // Check achievements after business level change
		return true
	}
	return false
}

// BuyUpgradeCard purchases a specific upgrade item by its index if balance is sufficient.
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
		e.applyModifiers() // Recalculate business modifiers after new upgrade purchased
		e.checkAchievements() // Check achievements
		return true
	}
	return false
}

// BuyManager hires a specific manager by its index if balance is sufficient.
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

// applyModifiers recalculates and applies modifiers from upgrades and achievements.
// Note: Caller must hold Lock on mu.
func (e *Engine) applyModifiers() {
	revMults := make(map[string]float64)
	speedMults := make(map[string]float64)

	for _, b := range e.businesses {
		revMults[b.ID] = 1.0
		speedMults[b.ID] = 1.0
	}

	// 1. Apply purchased upgrades
	for _, upg := range e.upgrades {
		if upg.IsPurchased {
			revMults[upg.TargetBusinessID] *= upg.Effect.RevenueMultiplier
			speedMults[upg.TargetBusinessID] *= upg.Effect.SpeedMultiplier
		}
	}

	// 2. Apply unlocked achievements bonuses
	for _, ach := range e.achievements {
		if ach.IsUnlocked {
			if ach.TargetBusinessID == "" {
				// Global bonus
				for _, b := range e.businesses {
					revMults[b.ID] *= (1.0 + ach.BonusMultiplier)
				}
			} else {
				// Specific business line bonus
				revMults[ach.TargetBusinessID] *= (1.0 + ach.BonusMultiplier)
			}
		}
	}

	// 3. Apply Angel Investors bonus (+5% global revenue per investor)
	if e.angels > 0 {
		angelMultiplier := 1.0 + float64(e.angels)*0.05
		for _, b := range e.businesses {
			revMults[b.ID] *= angelMultiplier
		}
	}

	// 4. Apply Super Boost (2x Global Speed if active)
	boostMult := 1.0
	if e.boostDuration > 0 {
		boostMult = 2.0
	}

	// 5. Apply active Random Event if any
	if e.activeEvent != nil {
		evt := e.activeEvent
		if evt.EffectType == "revenue" {
			if evt.TargetBizID == "" {
				for _, b := range e.businesses {
					revMults[b.ID] *= evt.Multiplier
				}
			} else {
				revMults[evt.TargetBizID] *= evt.Multiplier
			}
		} else if evt.EffectType == "speed" {
			if evt.TargetBizID == "" {
				for _, b := range e.businesses {
					speedMults[b.ID] *= evt.Multiplier
				}
			} else {
				speedMults[evt.TargetBizID] *= evt.Multiplier
			}
		}
	}

	// Apply to each business
	for _, b := range e.businesses {
		b.SetModifiers(revMults[b.ID], speedMults[b.ID]*boostMult)
	}
}

// applyAutomation applies automation based on hired managers.
// Note: Caller must hold Lock on mu.
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

// checkAchievements checks achievement conditions and unlocks them if met.
// Note: Caller must hold Lock on mu.
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

// TriggerProduction triggers a manual production start for non-automated businesses.
func (e *Engine) TriggerProduction(idx int) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if idx < 0 || idx >= len(e.businesses) {
		return false
	}

	return e.businesses[idx].StartProduction()
}

// ExportState exports the current engine state to a SaveState structure.
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
		BoostDurationNs:  int64(e.boostDuration),
		Timestamp:        time.Now(),
	}
}

// ImportState loads the save state into the engine and calculates offline earnings with NTP anti-cheat protection.
// Returns the total offline earnings gathered (capped at 12 hours offline).
func (e *Engine) ImportState(state *save.SaveState) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Restore wallet balance
	e.wallet = economy.NewWallet(state.Balance)

	// 2. Restore businesses state
	bizStateMap := make(map[string]save.BizState)
	for _, bs := range state.Businesses {
		bizStateMap[bs.ID] = bs
	}

	for _, b := range e.businesses {
		if bs, ok := bizStateMap[b.ID]; ok {
			b.LoadState(bs.Level, bs.IsActive, bs.ProgressNs)
		}
	}

	// 3. Restore upgrades state
	purchasedUpgMap := make(map[string]bool)
	for _, id := range state.Upgrades {
		purchasedUpgMap[id] = true
	}
	for _, upg := range e.upgrades {
		upg.IsPurchased = purchasedUpgMap[upg.ID]
	}

	// 4. Restore managers state
	hiredMgrMap := make(map[string]bool)
	for _, id := range state.Managers {
		hiredMgrMap[id] = true
	}
	for _, m := range e.managers {
		m.IsHired = hiredMgrMap[m.ID]
	}
	e.applyAutomation()

	// 5. Restore achievements state
	unlockedAchMap := make(map[string]bool)
	for _, id := range state.Achievements {
		unlockedAchMap[id] = true
	}
	for _, ach := range e.achievements {
		ach.IsUnlocked = unlockedAchMap[ach.ID]
	}

	// 6. Restore lifetime earnings and investors (backward compatible)
	e.lifetimeEarnings = state.LifetimeEarnings
	if e.lifetimeEarnings < e.wallet.Balance() {
		e.lifetimeEarnings = e.wallet.Balance()
	}
	e.angels = state.Angels
	e.boostDuration = time.Duration(state.BoostDurationNs)

	// 7. Recalculate modifiers
	e.applyModifiers()

	// 8. Calculate offline earnings (max 12 hours) with anti-cheat protection
	now := time.Now()
	offlineDuration := now.Sub(state.Timestamp)

	// Protection 1: Use NTP if connected
	ntpTime, ntpErr := gametime.GetNetworkTime("pool.ntp.org", 1500*time.Millisecond)
	if ntpErr == nil {
		// Validate if local time is accurate (5 minutes tolerance)
		isLocalValid := gametime.IsSystemTimeValid(now, ntpTime, 5*time.Minute)
		if !isLocalValid {
			// Local clock is manipulated, force use trusted NTP clock diff!
			offlineDuration = ntpTime.Sub(state.Timestamp)
		}

		// Validate if NTP time is before the last saved timestamp (clock rewind cheat)
		if ntpTime.Before(state.Timestamp) {
			offlineDuration = 0
		}
	} else {
		// Offline fallback to local time
		// Protection 2: Detect if local clock was rewound behind the last saved timestamp
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

	// Reset lastTick to current time so offline duration is not processed twice
	e.lastTick = now

	return offlineRevenue
}

// GetTotalGPS returns the total gain per second (GPS) from all automated businesses owned by the player.
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





