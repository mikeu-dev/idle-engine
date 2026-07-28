package ebiten

import (
	"fmt"
	"image/color"
	"idle-engine/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

// Update memproses input dan memperbarui engine.
func (g *Game) Update() error {
	g.engine.Update()

	// Kontrol untuk Trigger Produksi manual (Bisnis 0 - Lemonade Stand)
	if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) {
		g.engine.TriggerProduction(0)
	}

	// Kontrol untuk Upgrade Bisnis
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.engine.BuyUpgrade(0) // Upgrade Lemonade
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.engine.BuyUpgrade(1) // Upgrade Newspaper
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.engine.BuyUpgrade(2) // Upgrade Car Wash
	}

	return nil
}

// Draw menggambar representasi visual game ke layar.
func (g *Game) Draw(screen *ebiten.Image) {
	// Background color (Slate dark)
	screen.Fill(color.RGBA{R: 30, G: 30, B: 46, A: 255})

	wallet := g.engine.GetWallet()
	balance := wallet.Balance()

	// Header
	ebitenutil.DebugPrintAt(screen, "=== IDLE ENGINE SANDBOX ===", 20, 15)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SALDO: %.2f POIN", balance), 20, 35)

	// Gambar Lini Bisnis
	businesses := g.engine.GetBusinesses()
	for i, b := range businesses {
		y := 80 + i*115

		// Judul Bisnis & Level
		var statusText string
		if b.IsOwned() {
			statusText = fmt.Sprintf("Lv. %d", b.GetLevel())
		} else {
			statusText = "Belum Dimiliki"
		}
		
		businessTitle := fmt.Sprintf("%d. %s (%s)", i+1, b.Name, statusText)
		ebitenutil.DebugPrintAt(screen, businessTitle, 20, y)

		// Detail Produksi & Biaya Upgrade
		var detailText string
		var actionText string

		cost := b.Cost()
		canAfford := wallet.CanAfford(cost)
		
		// Warna teks indikator kemampuan membeli
		var costColor string
		if canAfford {
			costColor = "Bisa Beli"
		} else {
			costColor = "Saldo Kurang"
		}

		// Konfigurasi tombol shortcut
		var keyName string
		switch i {
		case 0:
			keyName = "Q"
		case 1:
			keyName = "W"
		case 2:
			keyName = "E"
		}

		if b.IsOwned() {
			detailText = fmt.Sprintf("Penghasilan: %.2f Poin / %s", b.Income(), b.Duration)
		} else {
			detailText = fmt.Sprintf("Penghasilan Awal: %.2f Poin / %s", b.BaseIncome, b.Duration)
		}

		if b.IsAutomated {
			actionText = fmt.Sprintf("Upgrade: Tekan [%s] (Biaya: %.2f) [%s]", keyName, cost, costColor)
		} else {
			actionText = fmt.Sprintf("Upgrade: Tekan [%s] (Biaya: %.2f) [%s] | Mulai: Tekan [%d]", keyName, cost, costColor, i+1)
		}

		ebitenutil.DebugPrintAt(screen, detailText, 20, y+20)
		ebitenutil.DebugPrintAt(screen, actionText, 20, y+38)

		// Progress Bar
		barX := 20
		barY := y + 58
		barW := 600
		barH := 12

		// Background Progress Bar (grayish-blue)
		vector.DrawFilledRect(screen, float32(barX), float32(barY), float32(barW), float32(barH), color.RGBA{R: 45, G: 45, B: 60, A: 255}, false)

		// Fill Progress Bar
		pct := b.GetProgressPercent()
		if pct > 0 {
			// Jika bisnis otomatis, gunakan warna hijau. Jika manual, gunakan warna biru muda.
			var barColor color.RGBA
			if b.IsAutomated {
				barColor = color.RGBA{R: 166, G: 227, B: 161, A: 255} // Pastel Green
			} else {
				barColor = color.RGBA{R: 137, G: 180, B: 250, A: 255} // Pastel Blue
			}
			vector.DrawFilledRect(screen, float32(barX), float32(barY), float32(float64(barW)*pct), float32(barH), barColor, false)
		}
	}

	// Footer / Petunjuk
	ebitenutil.DebugPrintAt(screen, "Petunjuk: Bisnis otomatis (Newspaper & Car Wash) akan langsung berproduksi setelah dibeli.", 20, 440)
}

// Layout mengembalikan ukuran layar game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}
