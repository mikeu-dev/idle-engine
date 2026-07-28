package config

import (
	"os"
	"testing"
)

func TestLoadConfigSuccess(t *testing.T) {
	// Buat berkas YAML sementara untuk testing
	yamlContent := `
businesses:
  - id: test_biz
    name: Test Business
    base_cost: 10.0
    cost_multiplier: 1.2
    base_income: 2.0
    duration_ms: 1000
    is_automated: true
upgrades:
  - id: test_upg
    name: Test Upgrade
    description: Doubler
    cost: 50.0
    target_business_id: test_biz
    revenue_multiplier: 2.0
    speed_multiplier: 1.0
managers:
  - id: test_mgr
    name: Test Manager
    description: Automator
    cost: 100.0
    target_business_id: test_biz
achievements:
  - id: test_ach
    name: Test Achievement
    description: Unlocked
    condition_type: balance
    target_business_id: ""
    target_value: 1000.0
    bonus_multiplier: 0.1
`
	tmpfile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Muat konfigurasi dari file sementara
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verifikasi lini bisnis
	if len(cfg.Businesses) != 1 || cfg.Businesses[0].ID != "test_biz" || cfg.Businesses[0].BaseCost != 10.0 {
		t.Errorf("business config parsed incorrectly: %+v", cfg.Businesses)
	}

	// Verifikasi upgrade
	if len(cfg.Upgrades) != 1 || cfg.Upgrades[0].ID != "test_upg" || cfg.Upgrades[0].RevenueMultiplier != 2.0 {
		t.Errorf("upgrade config parsed incorrectly: %+v", cfg.Upgrades)
	}

	// Verifikasi manager
	if len(cfg.Managers) != 1 || cfg.Managers[0].ID != "test_mgr" || cfg.Managers[0].Cost != 100.0 {
		t.Errorf("manager config parsed incorrectly: %+v", cfg.Managers)
	}

	// Verifikasi achievement
	if len(cfg.Achievements) != 1 || cfg.Achievements[0].ID != "test_ach" || cfg.Achievements[0].BonusMultiplier != 0.1 {
		t.Errorf("achievement config parsed incorrectly: %+v", cfg.Achievements)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("non_existent_file.yaml")
	if err == nil {
		t.Error("expected error loading non-existent file, got nil")
	}
}
