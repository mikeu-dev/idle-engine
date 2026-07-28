package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// BizConfig represents the initial configuration of a business line.
type BizConfig struct {
	ID             string  `yaml:"id"`
	Name           string  `yaml:"name"`
	BaseCost       float64 `yaml:"base_cost"`
	CostMultiplier float64 `yaml:"cost_multiplier"`
	BaseIncome     float64 `yaml:"base_income"`
	DurationMs     int     `yaml:"duration_ms"`
	IsAutomated    bool    `yaml:"is_automated"`
}

// UpgradeConfig represents the configuration of an upgrade item.
type UpgradeConfig struct {
	ID                string  `yaml:"id"`
	Name              string  `yaml:"name"`
	Description       string  `yaml:"description"`
	Cost              float64 `yaml:"cost"`
	TargetBusinessID  string  `yaml:"target_business_id"`
	RevenueMultiplier float64 `yaml:"revenue_multiplier"`
	SpeedMultiplier   float64 `yaml:"speed_multiplier"`
}

// ManagerConfig represents the configuration of an automation manager.
type ManagerConfig struct {
	ID               string  `yaml:"id"`
	Name             string  `yaml:"name"`
	Description      string  `yaml:"description"`
	Cost             float64 `yaml:"cost"`
	TargetBusinessID string  `yaml:"target_business_id"`
}

// AchievementConfig represents the configuration of an achievement requirement.
type AchievementConfig struct {
	ID               string  `yaml:"id"`
	Name             string  `yaml:"name"`
	Description      string  `yaml:"description"`
	ConditionType    string  `yaml:"condition_type"`
	TargetBusinessID string  `yaml:"target_business_id"`
	TargetValue      float64 `yaml:"target_value"`
	BonusMultiplier  float64 `yaml:"bonus_multiplier"`
}

// GameConfig holds the entire parsed game configurations.
type GameConfig struct {
	Businesses   []BizConfig         `yaml:"businesses"`
	Upgrades     []UpgradeConfig     `yaml:"upgrades"`
	Managers     []ManagerConfig     `yaml:"managers"`
	Achievements []AchievementConfig `yaml:"achievements"`
}

// LoadConfig loads the YAML configuration file from the specified path.
func LoadConfig(filepath string) (*GameConfig, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg GameConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
