package upgrade

import (
	"idle-engine/internal/domain/business"
	"idle-engine/internal/domain/modifier"
	"testing"
	"time"
)

func TestUpgradeAndModifiers(t *testing.T) {
		// Create manual business
	b := business.NewBusiness("lemonade", "Lemonade Stand", 10.0, 1.15, 2.0, time.Second, false)
	b.Upgrade() // level 1

	// Before upgrade: income = 2.0, duration = 1s
	if b.Income() != 2.0 {
		t.Errorf("expected income 2.0, got %f", b.Income())
	}

	// Create modifier: 2x revenue, 2x speed
	mod := modifier.NewModifier(2.0, 2.0)
	
	// Create upgrade item
	upg := NewUpgrade("lemon_pitcher", "Lemon Pitcher", "Double Lemonade revenue and speed", 15.0, "lemonade", mod)
	
	if upg.IsPurchased {
		t.Error("expected upgrade to be unpurchased initially")
	}

	// Purchase upgrade
	upg.Purchase()
	if !upg.IsPurchased {
		t.Error("expected upgrade to be purchased")
	}

	// Apply modifiers to business
	b.SetModifiers(upg.Effect.RevenueMultiplier, upg.Effect.SpeedMultiplier)

	// After modifier: income = 2.0 * 2.0 = 4.0
	if b.Income() != 4.0 {
		t.Errorf("expected income 4.0 after modifier, got %f", b.Income())
	}

	// Start production
	b.StartProduction()

	// Update with 500ms (speedMultiplier = 2x, so effective duration is 1s / 2 = 500ms)
	// Thus 500ms should complete the production
	revenue := b.Update(500 * time.Millisecond)
	if revenue != 4.0 {
		t.Errorf("expected revenue 4.0 after 500ms (speed 2x), got %f", revenue)
	}
	if b.GetIsActive() {
		t.Error("expected manual business to deactivate after completion")
	}
}
