package main

import (
	"log"
	"idle-engine/engine"
	presentation "idle-engine/internal/presentation/ebiten"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Inisialisasi logic engine
	eng := engine.NewEngine()

	// Inisialisasi presentasi ebiten
	game := presentation.NewGame(eng)

	// Setup jendela permainan
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Idle Engine Sandbox")

	log.Println("Memulai Idle Engine Sandbox...")

	// Jalankan game
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
