package engine

import (
	"testing"
	"time"
)

func TestRandomEventTriggerAndExpiration(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	// Upgrade business so GPS > 0
	e.GetBusinesses()[0].Upgrade() // Lemonade Stand level 1

	// Verify no active event initially
	evt, _ := e.GetActiveEvent()
	if evt != nil {
		t.Errorf("expected no active event initially, got %v", evt.ID)
	}

	// Trigger "summer" event (3x Lemonade Stand revenue)
	success := e.TriggerEventByID("summer")
	if !success {
		t.Fatal("expected summer event trigger to succeed")
	}

	evt, dur := e.GetActiveEvent()
	if evt == nil || evt.ID != "summer" {
		t.Errorf("expected active event ID to be 'summer', got %v", evt)
	}
	if dur != 30*time.Second {
		t.Errorf("expected event duration to be 30s, got %v", dur)
	}

	// Check Lemonade Stand income (BaseIncome = 1.0, level = 1, mult = 3.0 -> Income = 3.0)
	lemonade := e.GetBusinesses()[0]
	if lemonade.Income() != 3.0 {
		t.Errorf("expected lemonade income to be 3.0 under summer event, got %f", lemonade.Income())
	}

	// Run Update to simulate event expiration (30 seconds passed)
	e.lastTick = time.Now()
	e.Update()
	e.lastTick = e.lastTick.Add(-31 * time.Second) // Modify lastTick to simulate 31 seconds elapsed
	e.Update()

	// Verify event has expired
	evt, _ = e.GetActiveEvent()
	if evt != nil {
		t.Errorf("expected active event to expire, but still got %v", evt.ID)
	}

	// Verify Lemonade Stand income reverted to normal (1.0)
	if lemonade.Income() != 1.0 {
		t.Errorf("expected lemonade income to revert to 1.0, got %f", lemonade.Income())
	}
}

func TestRandomEventGlobalSpeedDebuff(t *testing.T) {
	e := NewEngineWithConfig("non_existent_config.yaml")

	news := e.GetBusinesses()[1]
	news.Upgrade() // Newspaper level 1

	// Initial GPS = 4.0 / 3.0s = 1.333333
	initialGPS := news.GetGPS()

	// Trigger power outage (power_outage) -> global speed reduced to 0.5x
	success := e.TriggerEventByID("power_outage")
	if !success {
		t.Fatal("expected power_outage event trigger to succeed")
	}

	// New GPS = 4.0 / (3.0s / 0.5) = 4.0 / 6.0s = 0.666667
	expectedGPS := initialGPS * 0.5
	diff := news.GetGPS() - expectedGPS
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("expected debuffed GPS to be around %f, got %f", expectedGPS, news.GetGPS())
	}
}
