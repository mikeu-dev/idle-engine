package business

import (
	"math"
	"sync"
	"time"
)

// Business merepresentasikan entitas lini bisnis penghasil uang.
type Business struct {
	mu             sync.RWMutex
	ID             string
	Name           string
	Level          int
	BaseCost       float64
	CostMultiplier float64
	BaseIncome     float64
	Duration       time.Duration
	Progress       time.Duration
	IsAutomated    bool
	IsActive       bool
}

// NewBusiness membuat instance Business baru.
func NewBusiness(id, name string, baseCost, costMultiplier, baseIncome float64, duration time.Duration, isAutomated bool) *Business {
	return &Business{
		ID:             id,
		Name:           name,
		Level:          0, // Level awal 0 (belum dibeli)
		BaseCost:       baseCost,
		CostMultiplier: costMultiplier,
		BaseIncome:     baseIncome,
		Duration:       duration,
		IsAutomated:    isAutomated,
	}
}

// Cost mengembalikan biaya untuk membeli atau meng-upgrade bisnis ke level berikutnya.
func (b *Business) Cost() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.BaseCost * math.Pow(b.CostMultiplier, float64(b.Level))
}

// Income mengembalikan pendapatan teoritis per siklus untuk level saat ini.
func (b *Business) Income() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.BaseIncome * float64(b.Level)
}

// Upgrade meningkatkan level bisnis.
func (b *Business) Upgrade() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Level++
	
	// Jika otomatis dan baru dibeli pertama kali, langsung aktifkan produksi
	if b.Level == 1 && b.IsAutomated {
		b.IsActive = true
		b.Progress = 0
	}
}

// StartProduction memulai siklus produksi untuk bisnis manual.
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

// Update memajukan siklus waktu bisnis berdasarkan delta durasi.
// Mengembalikan jumlah total pendapatan yang dihasilkan selama delta waktu tersebut.
func (b *Business) Update(delta time.Duration) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Jika belum dibeli atau tidak aktif berproduksi, tidak menghasilkan apa-apa
	if b.Level == 0 || !b.IsActive {
		return 0
	}

	revenue := 0.0
	b.Progress += delta

	if b.IsAutomated {
		// Untuk bisnis otomatis, kumpulkan pendapatan berulang secara efisien (O(1)) jika delta waktu besar
		numCycles := int64(b.Progress / b.Duration)
		if numCycles > 0 {
			revenue += b.BaseIncome * float64(b.Level) * float64(numCycles)
			b.Progress %= b.Duration
		}
	} else {
		// Untuk bisnis manual, selesaikan maksimal satu siklus dan matikan aktivitas
		if b.Progress >= b.Duration {
			revenue = b.BaseIncome * float64(b.Level)
			b.Progress = 0
			b.IsActive = false
		}
	}

	return revenue
}

// GetProgressPercent mengembalikan persentase proses produksi (0.0 hingga 1.0).
func (b *Business) GetProgressPercent() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Duration == 0 || !b.IsActive {
		return 0
	}
	pct := float64(b.Progress) / float64(b.Duration)
	if pct > 1.0 {
		return 1.0
	}
	return pct
}

// GetLevel mengembalikan level saat ini.
func (b *Business) GetLevel() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Level
}

// IsOwned memeriksa apakah bisnis ini sudah dibeli (level > 0).
func (b *Business) IsOwned() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Level > 0
}

// GetIsActive memeriksa apakah bisnis sedang berproduksi.
func (b *Business) GetIsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.IsActive
}

// LoadState memuat state penyimpanan ke dalam bisnis secara thread-safe.
func (b *Business) LoadState(level int, isActive bool, progress time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Level = level
	b.IsActive = isActive
	b.Progress = progress
}

// GetProgress mengembalikan progres waktu saat ini secara thread-safe.
func (b *Business) GetProgress() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Progress
}
