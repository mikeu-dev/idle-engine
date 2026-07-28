package manager

import (
	"idle-engine/internal/domain/business"
	"testing"
	"time"
)

func TestManagerHireAndAutomation(t *testing.T) {
		// Create manual business
	b := business.NewBusiness("lemonade", "Lemonade Stand", 10.0, 1.15, 2.0, time.Second, false)
	b.Upgrade() // level 1

	if b.IsAutomated {
		t.Error("expected lemonade to be manual initially")
	}

	// Create manager
	m := NewManager("lemon_mgr", "Lemonade Manager", "Automates Lemonade Stand", 100.0, "lemonade")

	if m.IsHired {
		t.Error("expected manager to be unhired initially")
	}

	// Hire manager
	m.Hire()
	if !m.IsHired {
		t.Error("expected manager to be hired")
	}

	// Apply automation to business
	b.SetAutomated(m.IsHired)

	if !b.IsAutomated {
		t.Error("expected business to become automated after hiring manager")
	}

	// Once automation is active on level > 0, business should automatically start active production
	if !b.GetIsActive() {
		t.Error("expected automated business to automatically start active production")
	}
}
