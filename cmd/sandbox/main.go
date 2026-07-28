package main

import (
	"log"
	"os"
	"idle-engine/engine"
	"idle-engine/internal/core/save"
	presentation "idle-engine/internal/presentation/ebiten"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Initialize logic engine
	eng := engine.NewEngine()

	// Try to load save data at startup
	var offlineRevenue float64
	saveFile := "savegame.json"
	if _, err := os.Stat(saveFile); err == nil {
		log.Println("Found savegame.json. Loading progress...")
		state, err := save.LoadFromFile(saveFile)
		if err != nil {
			log.Printf("Failed to load save file: %v", err)
		} else {
			offlineRevenue = eng.ImportState(state)
			log.Printf("Progress loaded successfully! Offline earnings: %.2f Points", offlineRevenue)
		}
	} else {
		log.Println("savegame.json not found. Starting from scratch.")
	}

	// Initialize Ebitengine presentation
	game := presentation.NewGame(eng)
	if offlineRevenue > 0 {
		game.SetOfflineNotification(offlineRevenue)
	}

	// Ensure game automatically saves progress when exiting normally
	defer func() {
		log.Println("Saving game progress before exit...")
		state := eng.ExportState()
		if err := save.SaveToFile(saveFile, state); err != nil {
			log.Printf("Failed to save game progress: %v", err)
		} else {
			log.Println("Game progress successfully saved to savegame.json!")
		}
	}()

	// Setup game window
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Idle Engine Sandbox")

	log.Println("Starting Idle Engine Sandbox...")

	// Run game
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

