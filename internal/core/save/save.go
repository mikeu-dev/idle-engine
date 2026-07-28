package save

import (
	"encoding/json"
	"os"
	"time"
)

// BizState holds save data for each business line.
type BizState struct {
	ID         string        `json:"id"`
	Level      int           `json:"level"`
	IsActive   bool          `json:"is_active"`
	ProgressNs time.Duration `json:"progress_ns"`
}

// SaveState is the main data structure for the game save file.
type SaveState struct {
	Balance          float64    `json:"balance"`
	Businesses       []BizState `json:"businesses"`
	Upgrades         []string   `json:"upgrades"`
	Managers         []string   `json:"managers"`
	Achievements     []string   `json:"achievements"`
	LifetimeEarnings float64    `json:"lifetime_earnings"`
	Angels           int        `json:"angels"`
	BoostDurationNs  int64      `json:"boost_duration_ns"`
	Timestamp        time.Time  `json:"timestamp"`
}

const xorKey byte = 0xAB

// SaveToFile writes the SaveState data to a local file using simple XOR encryption.
func SaveToFile(filename string, state *SaveState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Simple XOR encryption
	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ xorKey
	}

	return os.WriteFile(filename, encrypted, 0644)
}

// LoadFromFile reads the local file and returns the SaveState (supports both encrypted and legacy raw JSON).
func LoadFromFile(filename string) (*SaveState, error) {
	encrypted, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Decrypt data
	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i++ {
		decrypted[i] = encrypted[i] ^ xorKey
	}

	var state SaveState
	// Try to unmarshal the decrypted data first
	if err := json.Unmarshal(decrypted, &state); err != nil {
		// If it fails (likely a legacy unencrypted save), attempt to load the raw data
		if errLegacy := json.Unmarshal(encrypted, &state); errLegacy != nil {
			return nil, err
		}
	}
	return &state, nil
}
