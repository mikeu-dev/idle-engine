package ebiten

import (
	"fmt"
	"image/color"
	"idle-engine/engine"
	"idle-engine/internal/core/save"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game mengimplementasikan interface ebiten.Game.
type Game struct {
	engine             *engine.Engine
	lastAutoSave       time.Time
	notification       string
	notificationExpiry time.Time
	activeTab          int // 0: Bisnis, 1: Upgrade, 2: Manager, 3: Achievements
}

// NewGame membuat instance Game baru dengan engine yang diberikan.
func NewGame(eng *engine.Engine) *Game {
	return &Game{
		engine:       eng,
		lastAutoSave: time.Now(),
		activeTab:    0,
	}
}

// SetOfflineNotification mengatur pesan notifikasi pendapatan offline saat game pertama dibuka.
func (g *Game) SetOfflineNotification(revenue float64) {
	g.notification = fmt.Sprintf("[KEMBALI! PENDAPATAN OFFLINE: +%.2f POIN]", revenue)
	g.notificationExpiry = time.Now().Add(5 * time.Second)
}

// Update memproses input dan memperbarui engine serta siklus simpan-muat.
func (g *Game) Update() error {
	g.engine.Update()

	// Navigasi perpindahan tab menggunakan tombol TAB
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.activeTab = (g.activeTab + 1) % 4
		g.showTabChangeNotification()
	}

	// Navigasi perpindahan tab menggunakan F1-F4 secara langsung
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.activeTab = 0
		g.showTabChangeNotification()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		g.activeTab = 1
		g.showTabChangeNotification()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.activeTab = 2
		g.showTabChangeNotification()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		g.activeTab = 3
		g.showTabChangeNotification()
	}

	// Kontrol Hotkey Tab Bisnis (activeTab == 0)
	if g.activeTab == 0 {
		if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) {
			g.engine.TriggerProduction(0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			g.engine.BuyUpgrade(0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.engine.BuyUpgrade(1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			g.engine.BuyUpgrade(2)
		}
	}

	// Kontrol Hotkey Tab Upgrade (activeTab == 1)
	if g.activeTab == 1 {
		boughtIdx := -1
		if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			boughtIdx = 0
		} else if inpututil.IsKeyJustPressed(ebiten.Key2) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad2) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			boughtIdx = 1
		} else if inpututil.IsKeyJustPressed(ebiten.Key3) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad3) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
			boughtIdx = 2
		}

		if boughtIdx != -1 {
			upg := g.engine.GetUpgrades()[boughtIdx]
			if upg.IsPurchased {
				g.notification = fmt.Sprintf("[%s SUDAH DIBELI!]", upg.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else if g.engine.BuyUpgradeCard(boughtIdx) {
				g.notification = fmt.Sprintf("[%s BERHASIL DIBELI!]", upg.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else {
				g.notification = fmt.Sprintf("[GAGAL MEMBELI %s (SALDO KURANG)]", upg.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			}
		}
	}

	// Kontrol Hotkey Tab Manager (activeTab == 2)
	if g.activeTab == 2 {
		hiredIdx := -1
		if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			hiredIdx = 0
		}

		if hiredIdx != -1 {
			m := g.engine.GetManagers()[hiredIdx]
			if m.IsHired {
				g.notification = fmt.Sprintf("[%s SUDAH DIAKTIFKAN!]", m.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else if g.engine.BuyManager(hiredIdx) {
				g.notification = fmt.Sprintf("[%s BERHASIL DIREKRUT!]", m.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else {
				g.notification = fmt.Sprintf("[GAGAL MEREKRUT %s (SALDO KURANG)]", m.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			}
		}
	}

	// Kontrol Hotkey Manual Save/Load (Berlaku global)
	saveFile := "savegame.json"
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		state := g.engine.ExportState()
		if err := save.SaveToFile(saveFile, state); err != nil {
			g.notification = "[GAGAL MENYIMPAN GAME!]"
		} else {
			g.notification = "[GAME BERHASIL DISIMPAN!]"
		}
		g.notificationExpiry = time.Now().Add(3 * time.Second)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		state, err := save.LoadFromFile(saveFile)
		if err != nil {
			g.notification = "[GAGAL MEMUAT GAME!]"
		} else {
			g.engine.ImportState(state)
			g.notification = "[GAME BERHASIL DIMUAT!]"
		}
		g.notificationExpiry = time.Now().Add(3 * time.Second)
	}

	// Auto-Save tiap 5 detik
	if time.Since(g.lastAutoSave) >= 5*time.Second {
		g.lastAutoSave = time.Now()
		state := g.engine.ExportState()
		_ = save.SaveToFile(saveFile, state)
	}

	return nil
}

func (g *Game) showTabChangeNotification() {
	tabName := map[int]string{
		0: "LIS LINI BISNIS",
		1: "PENINGKATAN (UPGRADE)",
		2: "MANAJER (OTOMATISASI)",
		3: "PENCAPAIAN (ACHIEVEMENTS)",
	}[g.activeTab]
	g.notification = fmt.Sprintf("[TAB AKTIF: %s]", tabName)
	g.notificationExpiry = time.Now().Add(1500 * time.Millisecond)
}

// Draw menggambar representasi visual game ke layar.
func (g *Game) Draw(screen *ebiten.Image) {
	// Background color (Slate dark)
	screen.Fill(color.RGBA{R: 30, G: 30, B: 46, A: 255})

	wallet := g.engine.GetWallet()
	balance := wallet.Balance()

	// Header
	ebitenutil.DebugPrintAt(screen, "=== IDLE ENGINE SANDBOX ===", 20, 12)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SALDO: %.2f POIN", balance), 20, 30)

	// Draw active notification if not expired
	if time.Now().Before(g.notificationExpiry) && g.notification != "" {
		ebitenutil.DebugPrintAt(screen, g.notification, 320, 12)
	}

	// Gambar Tab Header (4 Tab UI)
	tabs := []struct {
		tabIdx int
		title  string
		hotkey string
		color  color.RGBA
	}{
		{0, " BISNIS", "F1", color.RGBA{R: 137, G: 180, B: 250, A: 255}}, // Blue
		{1, "UPGRADE", "F2", color.RGBA{R: 166, G: 227, B: 161, A: 255}}, // Green
		{2, "MANAGER", "F3", color.RGBA{R: 249, G: 226, B: 175, A: 255}}, // Yellow
		{3, "ACHIEVE", "F4", color.RGBA{R: 245, G: 194, B: 231, A: 255}}, // Pink/Lavender
	}

	for i, t := range tabs {
		x := 20 + i*112
		var bg color.RGBA
		var txt string
		if g.activeTab == t.tabIdx {
			bg = t.color
			txt = fmt.Sprintf("*%s*", t.title)
		} else {
			bg = color.RGBA{R: 45, G: 45, B: 60, A: 255}
			txt = fmt.Sprintf("[%s]%s", t.hotkey, t.title)
		}
		vector.DrawFilledRect(screen, float32(x), 50, 102, 22, bg, false)
		
		ebitenutil.DebugPrintAt(screen, txt, x+5, 53)
	}

	ebitenutil.DebugPrintAt(screen, "(Tekan F1-F4 atau [TAB])", 475, 53)

	// TAB 1: BISNIS
	if g.activeTab == 0 {
		businesses := g.engine.GetBusinesses()
		for i, b := range businesses {
			y := 88 + i*112

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
				actionText = fmt.Sprintf("Upgrade: Tekan [%s] (Biaya: %.2f) [%s] (Otomatis)", keyName, cost, costColor)
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
				var barColor color.RGBA
				if b.IsAutomated {
					barColor = color.RGBA{R: 166, G: 227, B: 161, A: 255} // Pastel Green
				} else {
					barColor = color.RGBA{R: 137, G: 180, B: 250, A: 255} // Pastel Blue
				}
				vector.DrawFilledRect(screen, float32(barX), float32(barY), float32(float64(barW)*pct), float32(barH), barColor, false)
			}
		}

		ebitenutil.DebugPrintAt(screen, "Petunjuk: Bisnis otomatis akan langsung berproduksi secara mandiri setelah dibeli.", 20, 422)
	}

	// TAB 2: UPGRADE
	if g.activeTab == 1 {
		upgrades := g.engine.GetUpgrades()
		for i, upg := range upgrades {
			y := 88 + i*108

			// Draw card background (box)
			vector.DrawFilledRect(screen, 20, float32(y), 600, 95, color.RGBA{R: 45, G: 45, B: 60, A: 255}, false)

			titleText := fmt.Sprintf("[%d] %s", i+1, upg.Name)
			if upg.IsPurchased {
				titleText += " (TERBELI)"
			}
			ebitenutil.DebugPrintAt(screen, titleText, 35, y+10)
			ebitenutil.DebugPrintAt(screen, upg.Description, 35, y+30)

			var actionText string
			if upg.IsPurchased {
				actionText = "Stat: Peningkatan Aktif"
			} else {
				canAfford := wallet.CanAfford(upg.Cost)
				statusColor := "Saldo Kurang"
				if canAfford {
					statusColor = "Bisa Beli"
				}
				actionText = fmt.Sprintf("Beli: Tekan [%d] / [%s] (Biaya: %.2f) [%s]", i+1, map[int]string{0: "Q", 1: "W", 2: "E"}[i], upg.Cost, statusColor)
			}
			ebitenutil.DebugPrintAt(screen, actionText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Petunjuk: Peningkatan memodifikasi pendapatan atau kecepatan produksi lini bisnis terkait.", 20, 422)
	}

	// TAB 3: MANAGER
	if g.activeTab == 2 {
		managers := g.engine.GetManagers()
		for i, m := range managers {
			y := 88 + i*108

			// Draw card background (box)
			vector.DrawFilledRect(screen, 20, float32(y), 600, 95, color.RGBA{R: 45, G: 45, B: 60, A: 255}, false)

			titleText := fmt.Sprintf("[%d] %s", i+1, m.Name)
			if m.IsHired {
				titleText += " (DIAKTIFKAN)"
			}
			ebitenutil.DebugPrintAt(screen, titleText, 35, y+10)
			ebitenutil.DebugPrintAt(screen, m.Description, 35, y+30)

			var actionText string
			if m.IsHired {
				actionText = "Otomatisasi: Aktif"
			} else {
				canAfford := wallet.CanAfford(m.Cost)
				statusColor := "Saldo Kurang"
				if canAfford {
					statusColor = "Bisa Beli"
				}
				actionText = fmt.Sprintf("Pekerjakan: Tekan [%d] / [%s] (Biaya: %.2f) [%s]", i+1, map[int]string{0: "Q", 1: "W", 2: "E"}[i], m.Cost, statusColor)
			}
			ebitenutil.DebugPrintAt(screen, actionText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Petunjuk: Manajer berguna untuk mengotomatiskan Lini Bisnis manual agar berproduksi secara mandiri.", 20, 422)
	}

	// TAB 4: ACHIEVEMENTS
	if g.activeTab == 3 {
		achievements := g.engine.GetAchievements()
		for i, ach := range achievements {
			y := 88 + i*108

			// Draw card background (box)
			var boxColor color.RGBA
			if ach.IsUnlocked {
				boxColor = color.RGBA{R: 50, G: 70, B: 60, A: 255} // Light green tint box for unlocked
			} else {
				boxColor = color.RGBA{R: 45, G: 45, B: 60, A: 255} // Dark gray box for locked
			}
			vector.DrawFilledRect(screen, 20, float32(y), 600, 95, boxColor, false)

			titleText := ach.Name
			if ach.IsUnlocked {
				titleText += " (TERBUKA)"
			} else {
				titleText += " (TERKUNCI)"
			}
			ebitenutil.DebugPrintAt(screen, titleText, 35, y+10)
			ebitenutil.DebugPrintAt(screen, ach.Description, 35, y+30)

			var statusText string
			if ach.IsUnlocked {
				statusText = "Status: Bonus +10%/20% Modifikasi Aktif"
			} else {
				statusText = "Status: Kondisi belum terpenuhi"
			}
			ebitenutil.DebugPrintAt(screen, statusText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Petunjuk: Pencapaian terbuka secara otomatis jika kondisi terpenuhi, memberikan bonus pendapatan permanen.", 20, 422)
	}

	// Footer (Global)
	ebitenutil.DebugPrintAt(screen, "Fitur: [S] Simpan Manual | [L] Muat Manual | Auto-save aktif (5s)", 20, 445)
}

// Layout mengembalikan ukuran layar game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}
