package manager

import (
	"idle-engine/internal/domain/business"
	"testing"
	"time"
)

func TestManagerHireAndAutomation(t *testing.T) {
	// Buat bisnis manual
	b := business.NewBusiness("lemonade", "Lemonade Stand", 10.0, 1.15, 2.0, time.Second, false)
	b.Upgrade() // level 1

	if b.IsAutomated {
		t.Error("expected lemonade to be manual initially")
	}

	// Buat manager
	m := NewManager("lemon_mgr", "Lemonade Manager", "Automates Lemonade Stand", 100.0, "lemonade")

	if m.IsHired {
		t.Error("expected manager to be unhired initially")
	}

	// Pekerjakan manager
	m.Hire()
	if !m.IsHired {
		t.Error("expected manager to be hired")
	}

	// Terapkan otomatisasi ke bisnis
	b.SetAutomated(m.IsHired)

	if !b.IsAutomated {
		t.Error("expected business to become automated after hiring manager")
	}

	// Setelah otomatisasi aktif pada level > 0, bisnis harus langsung aktif berproduksi
	if !b.GetIsActive() {
		t.Error("expected automated business to automatically start active production")
	}
}
