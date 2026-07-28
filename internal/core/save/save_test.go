package save

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_savegame.json")

	now := time.Now().UTC()
	originalState := &SaveState{
		Balance: 250.50,
		Businesses: []BizState{
			{ID: "lemonade", Level: 5, IsActive: true, ProgressNs: 500 * time.Millisecond},
			{ID: "newspaper", Level: 2, IsActive: false, ProgressNs: 0},
		},
		Timestamp: now,
	}

	// Test Save
	err := SaveToFile(tempFile, originalState)
	if err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Fatal("expected save file to exist")
	}

	// Test Load
	loadedState, err := LoadFromFile(tempFile)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// Verify balance
	if loadedState.Balance != originalState.Balance {
		t.Errorf("expected balance %f, got %f", originalState.Balance, loadedState.Balance)
	}

	// Verify timestamp (using Sub to avoid tiny float/serialization difference if any, though json unmarshal time.Time is precise)
	if !loadedState.Timestamp.Equal(originalState.Timestamp) {
		t.Errorf("expected timestamp %v, got %v", originalState.Timestamp, loadedState.Timestamp)
	}

	// Verify businesses
	if len(loadedState.Businesses) != len(originalState.Businesses) {
		t.Fatalf("expected %d businesses, got %d", len(originalState.Businesses), len(loadedState.Businesses))
	}

	for i, biz := range originalState.Businesses {
		loadedBiz := loadedState.Businesses[i]
		if loadedBiz.ID != biz.ID || loadedBiz.Level != biz.Level || loadedBiz.IsActive != biz.IsActive || loadedBiz.ProgressNs != biz.ProgressNs {
			t.Errorf("mismatch at index %d: expected %+v, got %+v", i, biz, loadedBiz)
		}
	}
}
