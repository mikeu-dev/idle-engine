package engine

import (
	"sync"
	"time"
)

// Engine merepresentasikan logic dan state dari game idle.
type Engine struct {
	mu           sync.RWMutex
	points       float64
	pointsPerSec float64
	lastTick     time.Time
}

// NewEngine membuat instance baru dari Engine.
func NewEngine() *Engine {
	return &Engine{
		points:       0.0,
		pointsPerSec: 1.0, // Mulai dengan menghasilkan 1 poin per detik
		lastTick:     time.Now(),
	}
}

// Update menghitung penambahan poin berdasarkan waktu yang berlalu.
func (e *Engine) Update() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(e.lastTick).Seconds()
	e.points += e.pointsPerSec * elapsed
	e.lastTick = now
}

// GetPoints mengembalikan total poin saat ini.
func (e *Engine) GetPoints() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.points
}

// GetPointsPerSec mengembalikan tingkat produksi poin saat ini.
func (e *Engine) GetPointsPerSec() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.pointsPerSec
}

// AddUpgrade menambah kecepatan produksi poin dengan biaya 10 poin.
func (e *Engine) AddUpgrade() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	cost := 10.0
	if e.points >= cost {
		e.points -= cost
		e.pointsPerSec += 1.0
		return true
	}
	return false
}
