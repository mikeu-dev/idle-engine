package business

import (
	"testing"
	"time"
)

func TestBusinessSpeedMilestones(t *testing.T) {
	// BaseIncome = 1.0, Duration = 2 seconds
	b := NewBusiness("test", "Test Business", 10.0, 1.15, 1.0, 2*time.Second, true)

	// Level 0: next milestone should be 25
	if b.GetNextMilestone() != 25 {
		t.Errorf("expected next milestone to be 25, got %d", b.GetNextMilestone())
	}

	// Level 24: next milestone still 25, multiplier 1.0
	b.UpgradeMany(24)
	if b.GetNextMilestone() != 25 {
		t.Errorf("expected next milestone to be 25, got %d", b.GetNextMilestone())
	}
	if b.getMilestoneSpeedMultiplier() != 1.0 {
		t.Errorf("expected multiplier to be 1.0, got %f", b.getMilestoneSpeedMultiplier())
	}

	// Level 25: next milestone 50, multiplier 2.0
	b.Upgrade()
	if b.GetNextMilestone() != 50 {
		t.Errorf("expected next milestone to be 50, got %d", b.GetNextMilestone())
	}
	if b.getMilestoneSpeedMultiplier() != 2.0 {
		t.Errorf("expected multiplier to be 2.0, got %f", b.getMilestoneSpeedMultiplier())
	}

	// Level 50: next milestone 100, multiplier 4.0 (2.0 * 2.0)
	b.UpgradeMany(25)
	if b.GetNextMilestone() != 100 {
		t.Errorf("expected next milestone to be 100, got %d", b.GetNextMilestone())
	}
	if b.getMilestoneSpeedMultiplier() != 4.0 {
		t.Errorf("expected multiplier to be 4.0, got %f", b.getMilestoneSpeedMultiplier())
	}
}

func TestBusinessGPSWithMilestones(t *testing.T) {
	// BaseIncome = 1.0, Duration = 1 second
	b := NewBusiness("test", "Test Business", 10.0, 1.15, 1.0, 1*time.Second, true)

	// Level 1: GPS = 1.0 / 1s = 1.0
	b.Upgrade()
	if b.GetGPS() != 1.0 {
		t.Errorf("expected GPS to be 1.0, got %f", b.GetGPS())
	}

	// Level 25: GPS = (1.0 * 25) / (1s / 2.0) = 25 / 0.5s = 50.0
	b.UpgradeMany(24)
	if b.GetGPS() != 50.0 {
		t.Errorf("expected GPS to be 50.0, got %f", b.GetGPS())
	}
}
