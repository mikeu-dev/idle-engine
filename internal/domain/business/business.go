package business

import (
	"math"
	"sync"
	"time"
)

// Business represents a money-producing business line entity.
type Business struct {
	mu                sync.RWMutex
	ID                string
	Name              string
	Level             int
	BaseCost          float64
	CostMultiplier    float64
	BaseIncome        float64
	Duration          time.Duration
	Progress          time.Duration
	IsAutomated       bool
	IsActive          bool
	revenueMultiplier float64
	speedMultiplier   float64
}

// NewBusiness creates a new Business instance.
func NewBusiness(id, name string, baseCost, costMultiplier, baseIncome float64, duration time.Duration, isAutomated bool) *Business {
	return &Business{
		ID:                id,
		Name:              name,
		Level:             0, // Level 0 initially (unpurchased)
		BaseCost:          baseCost,
		CostMultiplier:    costMultiplier,
		BaseIncome:        baseIncome,
		Duration:          duration,
		IsAutomated:       isAutomated,
		revenueMultiplier: 1.0,
		speedMultiplier:   1.0,
	}
}

// Cost returns the price to buy or upgrade the business to the next level.
func (b *Business) Cost() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.BaseCost * math.Pow(b.CostMultiplier, float64(b.Level))
}

// Income returns the theoretical revenue per cycle for the current level including multipliers.
func (b *Business) Income() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.BaseIncome * float64(b.Level) * b.revenueMultiplier
}

// Upgrade increases the business level.
func (b *Business) Upgrade() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Level++
	
	// If automated and purchased for the first time, start production immediately
	if b.Level == 1 && b.IsAutomated {
		b.IsActive = true
		b.Progress = 0
	}
}

// UpgradeMany increases the business level by multiple levels at once.
func (b *Business) UpgradeMany(levels int) {
	if levels <= 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	
	hadZeroLevel := (b.Level == 0)
	b.Level += levels
	
	// If automated and purchased for the first time, start production immediately
	if hadZeroLevel && b.Level >= 1 && b.IsAutomated {
		b.IsActive = true
		b.Progress = 0
	}
}

// StartProduction starts the production cycle for a manual business.
func (b *Business) StartProduction() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Level > 0 && !b.IsActive {
		b.IsActive = true
		b.Progress = 0
		return true
	}
	return false
}

// Update advances the business cycle time by the delta duration.
// Returns the total revenue generated during this time delta.
func (b *Business) Update(delta time.Duration) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If unowned or inactive, no revenue is generated
	if b.Level == 0 || !b.IsActive {
		return 0
	}

	revenue := 0.0
	b.Progress += delta

	// Calculate active duration based on speedMultiplier and milestone speedup
	speedMultiplierTotal := b.speedMultiplier * b.getMilestoneSpeedMultiplier()
	activeDuration := b.Duration
	if speedMultiplierTotal > 0 {
		activeDuration = time.Duration(float64(b.Duration) / speedMultiplierTotal)
	}
	if activeDuration == 0 {
		activeDuration = time.Nanosecond
	}

	incomePerCycle := b.BaseIncome * float64(b.Level) * b.revenueMultiplier

	if b.IsAutomated {
		// For automated business, efficiently collect repeating cycles (O(1)) for large deltas
		numCycles := int64(b.Progress / activeDuration)
		if numCycles > 0 {
			revenue += incomePerCycle * float64(numCycles)
			b.Progress %= activeDuration
		}
	} else {
		// For manual business, complete at most one cycle and deactivate
		if b.Progress >= activeDuration {
			revenue = incomePerCycle
			b.Progress = 0
			b.IsActive = false
		}
	}

	return revenue
}

// GetProgressPercent returns the production progress percentage (0.0 to 1.0).
func (b *Business) GetProgressPercent() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Duration == 0 || !b.IsActive {
		return 0
	}

	activeDuration := b.Duration
	if b.speedMultiplier > 0 {
		activeDuration = time.Duration(float64(b.Duration) / b.speedMultiplier)
	}
	if activeDuration == 0 {
		return 0
	}

	pct := float64(b.Progress) / float64(activeDuration)
	if pct > 1.0 {
		return 1.0
	}
	return pct
}

// SetModifiers updates revenue and speed multipliers in a thread-safe manner.
func (b *Business) SetModifiers(revMult, speedMult float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if revMult <= 0 {
		revMult = 1.0
	}
	if speedMult <= 0 {
		speedMult = 1.0
	}
	b.revenueMultiplier = revMult
	b.speedMultiplier = speedMult
}

// GetLevel returns the current level.
func (b *Business) GetLevel() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Level
}

// IsOwned checks if the business is purchased (level > 0).
func (b *Business) IsOwned() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Level > 0
}

// GetIsActive checks if the business is currently producing.
func (b *Business) GetIsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.IsActive
}

// LoadState loads save state data into the business in a thread-safe manner.
func (b *Business) LoadState(level int, isActive bool, progress time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Level = level
	b.IsActive = isActive
	b.Progress = progress
}

// GetProgress returns the current elapsed progress duration in a thread-safe manner.
func (b *Business) GetProgress() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Progress
}

// SetAutomated changes the business automation state in a thread-safe manner.
func (b *Business) SetAutomated(automated bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.IsAutomated = automated
	// If automated and level > 0, activate production immediately
	if b.IsAutomated && b.Level > 0 && !b.IsActive {
		b.IsActive = true
		b.Progress = 0
	}
}

// GetGPS returns the gain per second (GPS) projection for this business including milestone speedups.
func (b *Business) GetGPS() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.Level == 0 {
		return 0
	}
	speedMultiplierTotal := b.speedMultiplier * b.getMilestoneSpeedMultiplier()
	activeDuration := b.Duration
	if speedMultiplierTotal > 0 {
		activeDuration = time.Duration(float64(b.Duration) / speedMultiplierTotal)
	}
	if activeDuration.Seconds() <= 0 {
		return 0
	}
	return (b.BaseIncome * float64(b.Level) * b.revenueMultiplier) / activeDuration.Seconds()
}

// getMilestoneSpeedMultiplier calculates speed multiplier bonuses from level milestones.
// Each reached milestone doubles production speed.
// Note: Caller must hold RLock or Lock on b.mu.
func (b *Business) getMilestoneSpeedMultiplier() float64 {
	milestones := []int{25, 50, 100, 250, 500, 1000, 2500, 5000}
	mult := 1.0
	for _, m := range milestones {
		if b.Level >= m {
			mult *= 2.0
		}
	}
	return mult
}

// GetNextMilestone returns the next target milestone level based on current level in a thread-safe manner.
func (b *Business) GetNextMilestone() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	milestones := []int{25, 50, 100, 250, 500, 1000, 2500, 5000}
	for _, m := range milestones {
		if b.Level < m {
			return m
		}
	}
	// If exceeding the largest milestone, target the next multiple of 5000
	return ((b.Level / 5000) + 1) * 5000
}

