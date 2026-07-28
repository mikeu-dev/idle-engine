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
	Balance          float64    `json:"balance"`
	Businesses       []BizState `json:"businesses"`
	Upgrades         []string   `json:"upgrades"`
	Managers         []string   `json:"managers"`
	Achievements     []string   `json:"achievements"`
	LifetimeEarnings float64    `json:"lifetime_earnings"`
	Angels           int        `json:"angels"`
	Timestamp        time.Time  `json:"timestamp"`
}

const xorKey byte = 0xAB

// SaveToFile menulis data SaveState ke file lokal dalam bentuk terenkripsi XOR sederhana.
func SaveToFile(filename string, state *SaveState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Enkripsi XOR sederhana
	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ xorKey
	}

	return os.WriteFile(filename, encrypted, 0644)
}

// LoadFromFile membaca file lokal dan mengembalikan SaveState (mendukung file terenkripsi dan mentah legacy).
func LoadFromFile(filename string) (*SaveState, error) {
	encrypted, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Dekripsi data
	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i++ {
		decrypted[i] = encrypted[i] ^ xorKey
	}

	var state SaveState
	// Coba unmarshal data terdekripsi terlebih dahulu
	if err := json.Unmarshal(decrypted, &state); err != nil {
		// Jika gagal (kemungkinan karena file lama yang belum dienkripsi), coba muat data mentah asli
		if errLegacy := json.Unmarshal(encrypted, &state); errLegacy != nil {
			return nil, err
		}
	}
	return &state, nil
}
