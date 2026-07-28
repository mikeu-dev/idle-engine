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
	// Inisialisasi logic engine
	eng := engine.NewEngine()

	// Coba muat data simpanan saat startup
	var offlineRevenue float64
	saveFile := "savegame.json"
	if _, err := os.Stat(saveFile); err == nil {
		log.Println("Menemukan berkas savegame.json. Memuat progres...")
		state, err := save.LoadFromFile(saveFile)
		if err != nil {
			log.Printf("Gagal memuat berkas simpanan: %v", err)
		} else {
			offlineRevenue = eng.ImportState(state)
			log.Printf("Progres berhasil dimuat! Pendapatan offline: %.2f Poin", offlineRevenue)
		}
	} else {
		log.Println("Berkas savegame.json tidak ditemukan. Memulai dari awal.")
	}

	// Inisialisasi presentasi ebiten
	game := presentation.NewGame(eng)
	if offlineRevenue > 0 {
		game.SetOfflineNotification(offlineRevenue)
	}

	// Pastikan game otomatis menyimpan progres ketika keluar secara normal
	defer func() {
		log.Println("Menyimpan progres game sebelum keluar...")
		state := eng.ExportState()
		if err := save.SaveToFile(saveFile, state); err != nil {
			log.Printf("Gagal menyimpan progres game: %v", err)
		} else {
			log.Println("Progres game berhasil disimpan ke savegame.json!")
		}
	}()

	// Setup jendela permainan
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Idle Engine Sandbox")

	log.Println("Memulai Idle Engine Sandbox...")

	// Jalankan game
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

