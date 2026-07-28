package save

import (
	"encoding/json"
	"os"
	"time"
)

// BizState menampung data penyimpanan untuk setiap lini bisnis.
type BizState struct {
	ID         string        `json:"id"`
	Level      int           `json:"level"`
	IsActive   bool          `json:"is_active"`
	ProgressNs time.Duration `json:"progress_ns"`
}

// SaveState adalah struktur data utama untuk file penyimpanan game (JSON).
type SaveState struct {
	Balance    float64    `json:"balance"`
	Businesses []BizState `json:"businesses"`
	Timestamp  time.Time  `json:"timestamp"`
}

// SaveToFile menulis data SaveState ke file lokal dalam format JSON.
func SaveToFile(filename string, state *SaveState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile membaca file lokal dan mengembalikan SaveState.
func LoadFromFile(filename string) (*SaveState, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var state SaveState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
