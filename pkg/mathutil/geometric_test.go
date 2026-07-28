package mathutil

import (
	"math"
	"testing"
)

func TestCalculateMaxLevelsAffordable(t *testing.T) {
	// Case: Base Cost = 10, multiplier = 1.5, currentLevel = 0
	// Level 1: cost = 10 (total 10)
	// Level 2: cost = 15 (total 25)
	// Level 3: cost = 22.5 (total 47.5)

	// Test exact balance for 2 levels (25.0)
	levels, cost := CalculateMaxLevelsAffordable(10.0, 1.5, 0, 25.0)
	if levels != 2 {
		t.Errorf("expected 2 levels, got %d", levels)
	}
	if math.Abs(cost-25.0) > 1e-9 {
		t.Errorf("expected cost 25.0, got %f", cost)
	}

	// Test balance slightly below 25.0 (24.9)
	levels, cost = CalculateMaxLevelsAffordable(10.0, 1.5, 0, 24.9)
	if levels != 1 {
		t.Errorf("expected 1 level, got %d", levels)
	}
	if math.Abs(cost-10.0) > 1e-9 {
		t.Errorf("expected cost 10.0, got %f", cost)
	}

	// Test balance below base cost (3.0)
	levels, cost = CalculateMaxLevelsAffordable(10.0, 1.5, 0, 3.0)
	if levels != 0 || cost != 0.0 {
		t.Errorf("expected 0 levels and 0 cost, got %d and %f", levels, cost)
	}
}
