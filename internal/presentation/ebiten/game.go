package ebiten

import (
	"fmt"
	"image/color"
	"idle-engine/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Game mengimplementasikan interface ebiten.Game.
type Game struct {
	engine *engine.Engine
}

// NewGame membuat instance Game baru dengan engine yang diberikan.
func NewGame(eng *engine.Engine) *Game {
	return &Game{
		engine: eng,
	}
}

// Update dipanggil setiap frame untuk memperbarui state game.
func (g *Game) Update() error {
	// Perbarui logik engine
	g.engine.Update()

	// Jika tombol U ditekan, beli upgrade jika poin mencukupi
	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		g.engine.AddUpgrade()
	}

	return nil
}

// Draw menggambar visual game ke layar.
func (g *Game) Draw(screen *ebiten.Image) {
	// Warna latar belakang (dark theme)
	screen.Fill(color.RGBA{R: 30, G: 30, B: 46, A: 255})

	points := g.engine.GetPoints()
	pps := g.engine.GetPointsPerSec()

	msg := fmt.Sprintf(
		"=== IDLE ENGINE SANDBOX ===\n\n"+
			"Points: %.2f\n"+
			"Kecepatan Produksi: %.1f points/detik\n\n"+
			"Kontrol:\n"+
			"- Tekan [U] untuk beli Upgrade (Biaya: 10 points)\n",
		points, pps,
	)
	ebitenutil.DebugPrint(screen, msg)
}

// Layout mengembalikan ukuran layar permainan.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}
