package ebiten

import (
	"fmt"
	"image/color"
	"idle-engine/engine"
	"idle-engine/internal/core/save"
	"idle-engine/pkg/mathutil"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// FloatingText merepresentasikan teks mengambang kecil di layar.
type FloatingText struct {
	x, y      float32
	text      string
	alpha     float32
	createdAt time.Time
}

// UITheme mendefinisikan palet warna terintegrasi untuk tampilan game.
type UITheme struct {
	Name      string
	BgColor   color.RGBA
	CardBg    color.RGBA
	Blue      color.RGBA
	Green     color.RGBA
	Yellow    color.RGBA
	Pink      color.RGBA
	Orange    color.RGBA
}

// Game mengimplementasikan interface ebiten.Game.
type Game struct {
	engine                *engine.Engine
	lastAutoSave          time.Time
	notification          string
	notificationExpiry    time.Time
	activeTab             int // 0: Bisnis, 1: Upgrade, 2: Manager, 3: Achievements, 4: Investor
	prestigeConfirm       bool
	prestigeConfirmExpiry time.Time
	buyMaxMode            bool
	prevProgress          []time.Duration
	floatingTexts         []FloatingText
	activeThemeIndex      int
	themes                []UITheme
}

// NewGame membuat instance Game baru dengan engine yang diberikan.
func NewGame(eng *engine.Engine) *Game {
	themes := []UITheme{
		{
			Name:    "Catppuccin Mocha",
			BgColor: color.RGBA{R: 30, G: 30, B: 46, A: 255},
			CardBg:  color.RGBA{R: 45, G: 45, B: 60, A: 255},
			Blue:    color.RGBA{R: 137, G: 180, B: 250, A: 255},
			Green:   color.RGBA{R: 166, G: 227, B: 161, A: 255},
			Yellow:  color.RGBA{R: 249, G: 226, B: 175, A: 255},
			Pink:    color.RGBA{R: 245, G: 194, B: 231, A: 255},
			Orange:  color.RGBA{R: 250, G: 179, B: 135, A: 255},
		},
		{
			Name:    "Cyberpunk Neon",
			BgColor: color.RGBA{R: 10, G: 10, B: 20, A: 255},
			CardBg:  color.RGBA{R: 25, G: 15, B: 35, A: 255},
			Blue:    color.RGBA{R: 0, G: 240, B: 255, A: 255},
			Green:   color.RGBA{R: 57, G: 255, B: 20, A: 255},
			Yellow:  color.RGBA{R: 255, G: 255, B: 51, A: 255},
			Pink:    color.RGBA{R: 255, G: 0, B: 127, A: 255},
			Orange:  color.RGBA{R: 255, G: 110, B: 0, A: 255},
		},
		{
			Name:    "Nordic Frost",
			BgColor: color.RGBA{R: 46, G: 52, B: 64, A: 255},
			CardBg:  color.RGBA{R: 59, G: 66, B: 82, A: 255},
			Blue:    color.RGBA{R: 136, G: 192, B: 208, A: 255},
			Green:   color.RGBA{R: 163, G: 190, B: 140, A: 255},
			Yellow:  color.RGBA{R: 235, G: 203, B: 139, A: 255},
			Pink:    color.RGBA{R: 180, G: 142, B: 173, A: 255},
			Orange:  color.RGBA{R: 208, G: 135, B: 112, A: 255},
		},
	}

	return &Game{
		engine:           eng,
		lastAutoSave:     time.Now(),
		activeTab:        0,
		buyMaxMode:       false,
		themes:           themes,
		activeThemeIndex: 0,
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

	// Reset prestige confirm state jika sudah kedaluwarsa
	if g.prestigeConfirm && time.Now().After(g.prestigeConfirmExpiry) {
		g.prestigeConfirm = false
	}

	// Toggle Mode Beli dengan M (Berlaku Global)
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.buyMaxMode = !g.buyMaxMode
		if g.buyMaxMode {
			g.notification = "[MODE UPGRADE: BELI MAKSIMAL (MAX)]"
		} else {
			g.notification = "[MODE UPGRADE: BELI 1x]"
		}
		g.notificationExpiry = time.Now().Add(2 * time.Second)
	}

	// Toggle Tema Warna UI dengan T (Berlaku Global)
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.activeThemeIndex = (g.activeThemeIndex + 1) % len(g.themes)
		g.notification = fmt.Sprintf("[TEMA WARNA UI AKTIF: %s]", g.themes[g.activeThemeIndex].Name)
		g.notificationExpiry = time.Now().Add(2 * time.Second)
	}

	// Update Floating Texts
	var activeTexts []FloatingText
	now := time.Now()
	for _, ft := range g.floatingTexts {
		age := now.Sub(ft.createdAt)
		if age < 800*time.Millisecond {
			ft.y -= 0.6
			ft.alpha = 1.0 - float32(age.Milliseconds())/800.0
			activeTexts = append(activeTexts, ft)
		}
	}
	g.floatingTexts = activeTexts

	// Deteksi Penyelesaian Siklus Bisnis untuk Floating Text
	businesses := g.engine.GetBusinesses()
	if len(g.prevProgress) != len(businesses) {
		g.prevProgress = make([]time.Duration, len(businesses))
		for i, b := range businesses {
			g.prevProgress[i] = b.GetProgress()
		}
	}
	for i, b := range businesses {
		currProg := b.GetProgress()
		if b.IsOwned() && currProg < g.prevProgress[i] {
			yPos := float32(88 + i*112 + 58)
			g.spawnFloatingText(550, yPos, fmt.Sprintf("+%.2f Poin", b.Income()))
		}
		g.prevProgress[i] = currProg
	}

	// Navigasi perpindahan tab menggunakan tombol TAB
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.activeTab = (g.activeTab + 1) % 5
		g.showTabChangeNotification()
	}

	// Navigasi perpindahan tab menggunakan F1-F5 secara langsung
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
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		g.activeTab = 4
		g.showTabChangeNotification()
	}

	// Kontrol Hotkey Tab Bisnis (activeTab == 0)
	if g.activeTab == 0 {
		if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) {
			g.engine.TriggerProduction(0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			if g.buyMaxMode {
				g.engine.BuyUpgradeMax(0)
			} else {
				g.engine.BuyUpgrade(0)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyW) {
			if g.buyMaxMode {
				g.engine.BuyUpgradeMax(1)
			} else {
				g.engine.BuyUpgrade(1)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			if g.buyMaxMode {
				g.engine.BuyUpgradeMax(2)
			} else {
				g.engine.BuyUpgrade(2)
			}
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

		// Pemicu Booster Sementara
		if inpututil.IsKeyJustPressed(ebiten.KeyU) {
			if g.engine.TriggerSuperBoost() {
				g.notification = "[SUPER BOOST DIAKTIFKAN: KECEPATAN 2X!]"
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else {
				g.notification = "[GAGAL AKTIFKAN SUPER BOOST (BUTUH 50 POIN)]"
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyI) {
			if g.engine.TriggerTimeWarp() {
				g.notification = "[TIME WARP BERHASIL: INSTAN +1 JAM PENDAPATAN PASIF!]"
				g.notificationExpiry = time.Now().Add(3 * time.Second)
			} else {
				hasAuto := false
				for _, b := range g.engine.GetBusinesses() {
					if b.IsOwned() && b.IsAutomated {
						hasAuto = true
						break
					}
				}
				if !hasAuto {
					g.notification = "[GAGAL: TIME WARP MEMBUTUHKAN MINIMAL 1 BISNIS OTOMATIS]"
				} else {
					g.notification = "[GAGAL TIME WARP (BUTUH 150 POIN)]"
				}
				g.notificationExpiry = time.Now().Add(3 * time.Second)
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

	// Kontrol Hotkey Tab Investor / Prestige (activeTab == 4)
	if g.activeTab == 4 {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			claimable := g.engine.CalculateAngelsToClaim()
			if claimable <= 0 {
				g.notification = "[BELUM BISA PRESTIGE (TIDAK ADA INVESTOR UNTUK DIKLAIM)]"
				g.notificationExpiry = time.Now().Add(3 * time.Second)
				g.prestigeConfirm = false
			} else if g.prestigeConfirm && time.Now().Before(g.prestigeConfirmExpiry) {
				// Jalankan reset dan klaim
				if g.engine.ClaimPrestige() {
					g.notification = "[PRESTIGE BERHASIL! PROGRES DIRESET DENGAN BONUS INVESTOR BARU]"
					g.notificationExpiry = time.Now().Add(4 * time.Second)
				}
				g.prestigeConfirm = false
			} else {
				// Memasuki status konfirmasi
				g.prestigeConfirm = true
				g.prestigeConfirmExpiry = time.Now().Add(4 * time.Second)
				g.notification = "[TEKAN 'R' SEKALI LAGI UNTUK KONFIRMASI RESET & KLAIM!]"
				g.notificationExpiry = time.Now().Add(4 * time.Second)
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

func (g *Game) spawnFloatingText(x, y float32, text string) {
	g.floatingTexts = append(g.floatingTexts, FloatingText{
		x:         x,
		y:         y,
		text:      text,
		alpha:     1.0,
		createdAt: time.Now(),
	})
}

func (g *Game) showTabChangeNotification() {
	tabName := map[int]string{
		0: "LIS LINI BISNIS",
		1: "PENINGKATAN (UPGRADE)",
		2: "MANAJER (OTOMATISASI)",
		3: "PENCAPAIAN (ACHIEVEMENTS)",
		4: "INVESTOR MALAIKAT (PRESTIGE)",
	}[g.activeTab]
	g.notification = fmt.Sprintf("[TAB AKTIF: %s]", tabName)
	g.notificationExpiry = time.Now().Add(1500 * time.Millisecond)
}

// Draw menggambar representasi visual game ke layar.
func (g *Game) Draw(screen *ebiten.Image) {
	theme := g.themes[g.activeThemeIndex]

	// Background color dinamis sesuai tema aktif
	screen.Fill(theme.BgColor)

	wallet := g.engine.GetWallet()
	balance := wallet.Balance()

	// Header
	ebitenutil.DebugPrintAt(screen, "=== IDLE ENGINE SANDBOX ===", 20, 12)
	boostDur := g.engine.GetBoostDuration()
	var gpsStr string
	if boostDur > 0 {
		gpsStr = fmt.Sprintf("SALDO: %.2f POIN (+%.2f/dtk) [BOOST: %.1fs]", balance, g.engine.GetTotalGPS(), boostDur.Seconds())
	} else {
		gpsStr = fmt.Sprintf("SALDO: %.2f POIN (+%.2f/dtk)", balance, g.engine.GetTotalGPS())
	}
	ebitenutil.DebugPrintAt(screen, gpsStr, 20, 30)

	modeStr := "MODE BELI: [1x] (Tekan [M] untuk Maks)"
	if g.buyMaxMode {
		modeStr = "MODE BELI: [MAKS] (Tekan [M] untuk 1x)"
	}
	ebitenutil.DebugPrintAt(screen, modeStr, 340, 30)

	// Draw active notification if not expired
	if time.Now().Before(g.notificationExpiry) && g.notification != "" {
		ebitenutil.DebugPrintAt(screen, g.notification, 320, 12)
	}

	// Gambar Tab Header (5 Tab UI) dengan warna dinamis sesuai tema aktif
	tabs := []struct {
		tabIdx int
		title  string
		hotkey string
		color  color.RGBA
	}{
		{0, " BISNIS", "F1", theme.Blue},
		{1, "UPGRADE", "F2", theme.Green},
		{2, "MANAGER", "F3", theme.Yellow},
		{3, "ACHIEVE", "F4", theme.Pink},
		{4, "INVESTOR", "F5", theme.Orange},
	}

	for i, t := range tabs {
		x := 20 + i*90
		var bg color.RGBA
		var txt string
		if g.activeTab == t.tabIdx {
			bg = t.color
			txt = fmt.Sprintf("*%s*", t.title)
		} else {
			bg = theme.CardBg
			txt = fmt.Sprintf("[%s]%s", t.hotkey, t.title)
		}
		vector.DrawFilledRect(screen, float32(x), 50, 84, 22, bg, false)
		
		// Text color contrast
		ebitenutil.DebugPrintAt(screen, txt, x+3, 53)
	}

	ebitenutil.DebugPrintAt(screen, "(Tekan F1-F5 atau [TAB])", 475, 53)

	// TAB 1: BISNIS
	if g.activeTab == 0 {
		businesses := g.engine.GetBusinesses()
		for i, b := range businesses {
			y := 88 + i*112

			// Judul Bisnis & Level dengan Target Milestones Kecepatan
			var statusText string
			if b.IsOwned() {
				statusText = fmt.Sprintf("Lv. %d/%d (Akselerasi 2x)", b.GetLevel(), b.GetNextMilestone())
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
				detailText = fmt.Sprintf("Penghasilan: %.2f Poin / %s (+%.2f/dtk)", b.Income(), b.Duration, b.GetGPS())
			} else {
				detailText = fmt.Sprintf("Penghasilan Awal: %.2f Poin / %s", b.BaseIncome, b.Duration)
			}

			if g.buyMaxMode {
				levels, totalCost := mathutil.CalculateMaxLevelsAffordable(b.BaseCost, b.CostMultiplier, b.GetLevel(), balance)
				if levels > 0 {
					actionText = fmt.Sprintf("Beli Maks: Tekan [%s] (Dapat +%d Level, Biaya: %.2f) [Bisa Beli]", keyName, levels, totalCost)
				} else {
					actionText = fmt.Sprintf("Beli Maks: Tekan [%s] (Biaya: %.2f) [Saldo Kurang]", keyName, cost)
				}
			} else {
				if b.IsAutomated {
					actionText = fmt.Sprintf("Upgrade: Tekan [%s] (Biaya: %.2f) [%s] (Otomatis)", keyName, cost, costColor)
				} else {
					actionText = fmt.Sprintf("Upgrade: Tekan [%s] (Biaya: %.2f) [%s] | Mulai: Tekan [%d]", keyName, cost, costColor, i+1)
				}
			}

			ebitenutil.DebugPrintAt(screen, detailText, 20, y+20)
			ebitenutil.DebugPrintAt(screen, actionText, 20, y+38)

			// Progress Bar
			barX := 20
			barY := y + 58
			barW := 600
			barH := 12

			// Background Progress Bar (grayish-blue) dari tema aktif
			vector.DrawFilledRect(screen, float32(barX), float32(barY), float32(barW), float32(barH), theme.CardBg, false)

			// Fill Progress Bar
			pct := b.GetProgressPercent()
			if pct > 0 {
				var barColor color.RGBA
				if b.IsAutomated {
					barColor = theme.Green
				} else {
					barColor = theme.Blue
				}
				vector.DrawFilledRect(screen, float32(barX), float32(barY), float32(float64(barW)*pct), float32(barH), barColor, false)
			}
		}

		ebitenutil.DebugPrintAt(screen, "Petunjuk: Bisnis otomatis akan langsung berproduksi secara mandiri setelah dibeli.", 20, 422)
	}

	// TAB 2: UPGRADE
	if g.activeTab == 1 {
		// 1. Upgrade Cards di Sisi Kiri (x=20 sampai x=400)
		upgrades := g.engine.GetUpgrades()
		for i, upg := range upgrades {
			y := 88 + i*108

			// Draw card background (box lebar 380) dari tema aktif
			vector.DrawFilledRect(screen, 20, float32(y), 380, 95, theme.CardBg, false)

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
				actionText = fmt.Sprintf("Beli: Tekan [%s] (Biaya: %.2f) [%s]", map[int]string{0: "Q", 1: "W", 2: "E"}[i], upg.Cost, statusColor)
			}
			ebitenutil.DebugPrintAt(screen, actionText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Petunjuk: Peningkatan memodifikasi multiplier permanen.", 20, 422)

		// 2. Toko Booster Sementara di Sisi Kanan (x=420 sampai x=620)
		ebitenutil.DebugPrintAt(screen, "=== TOKO BOOSTER SEMENTARA ===", 420, 88)

		// Kartu Super Boost (y=110)
		vector.DrawFilledRect(screen, 420, 110, 200, 100, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "[U] SUPER BOOST", 435, 120)
		ebitenutil.DebugPrintAt(screen, "2x Kecepatan (30s)", 435, 140)
		ebitenutil.DebugPrintAt(screen, "Biaya: 50.00 Poin", 435, 160)
		boostStatus := "[Saldo Kurang]"
		if wallet.CanAfford(50.0) {
			boostStatus = "[Bisa Beli]"
		}
		ebitenutil.DebugPrintAt(screen, boostStatus, 435, 180)

		// Kartu Time Warp (y=230)
		vector.DrawFilledRect(screen, 420, 230, 200, 100, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "[I] TIME WARP", 435, 240)
		ebitenutil.DebugPrintAt(screen, "Instan +1 Jam Otomatis", 435, 260)
		ebitenutil.DebugPrintAt(screen, "Biaya: 150.00 Poin", 435, 280)
		warpStatus := "[Saldo Kurang]"
		if wallet.CanAfford(150.0) {
			warpStatus = "[Bisa Beli]"
		}
		ebitenutil.DebugPrintAt(screen, warpStatus, 435, 300)
	}

	// TAB 3: MANAGER
	if g.activeTab == 2 {
		managers := g.engine.GetManagers()
		for i, m := range managers {
			y := 88 + i*108

			// Draw card background (box) dari tema aktif
			vector.DrawFilledRect(screen, 20, float32(y), 600, 95, theme.CardBg, false)

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

			// Draw card background (box) dari tema aktif
			var boxColor color.RGBA
			if ach.IsUnlocked {
				boxColor = color.RGBA{
					R: uint8(float64(theme.Green.R)*0.4 + float64(theme.CardBg.R)*0.6),
					G: uint8(float64(theme.Green.G)*0.4 + float64(theme.CardBg.G)*0.6),
					B: uint8(float64(theme.Green.B)*0.4 + float64(theme.CardBg.B)*0.6),
					A: 255,
				}
			} else {
				boxColor = theme.CardBg
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

	// TAB 5: INVESTOR / PRESTIGE
	if g.activeTab == 4 {
		lifetime := g.engine.GetLifetimeEarnings()
		angels := g.engine.GetAngels()
		claimable := g.engine.CalculateAngelsToClaim()

		// 1. Statistik Card dari tema aktif
		vector.DrawFilledRect(screen, 20, 88, 600, 110, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "STATISTIK SEPANJANG MASA (LIFETIME STATS):", 35, 98)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Total Akumulasi Pendapatan : %.2f Poin", lifetime), 35, 120)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Angel Investors Dimiliki  : %d Investor", angels), 35, 142)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Bonus Pengali Pendapatan  : +%d%% Pendapatan (Global)", angels*5), 35, 164)

		// 2. Prestige Card dari tema aktif
		vector.DrawFilledRect(screen, 20, 218, 600, 130, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "RESET PRESTIS (INVESTASI MALAIKAT):", 35, 228)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Investor Baru untuk Diklaim : +%d Investor", claimable), 35, 250)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Efek Bonus Setelah Klaim    : +%d%% Pendapatan (Global)", (angels+claimable)*5), 35, 272)

		var prestigeText string
		if claimable <= 0 {
			prestigeText = "Tindakan: Belum ada investor baru untuk diklaim (Minimal 100 Poin sepanjang masa)."
		} else if g.prestigeConfirm {
			prestigeText = "[TEKAN 'R' SEKALI LAGI UNTUK KONFIRMASI RESET & KLAIM!]"
		} else {
			prestigeText = "Tindakan: Tekan [R] untuk memicu reset progres & klaim investor baru."
		}
		ebitenutil.DebugPrintAt(screen, prestigeText, 35, 305)

		// Keterangan
		ebitenutil.DebugPrintAt(screen, "Info: Melakukan Prestige akan mereset saldo, level bisnis, upgrade, dan manajer Anda.", 20, 365)
		ebitenutil.DebugPrintAt(screen, "Namun, Angel Investors memberikan +5% pendapatan permanen global. Pencapaian tidak direset.", 20, 385)
	}

	// Footer (Global) & Banner Event Acak Aktif
	activeEvt, eventDur := g.engine.GetActiveEvent()
	if activeEvt != nil {
		eventStr := fmt.Sprintf("[BERITA SELA] %s: %s (Sisa %.1fs)", activeEvt.Title, activeEvt.Description, eventDur.Seconds())
		ebitenutil.DebugPrintAt(screen, eventStr, 20, 435)
	}
	ebitenutil.DebugPrintAt(screen, "Fitur: [S] Simpan Manual | [L] Muat Manual | Auto-save aktif (5s)", 20, 455)

	// Render Floating Texts
	for _, ft := range g.floatingTexts {
		ebitenutil.DebugPrintAt(screen, ft.text, int(ft.x), int(ft.y))
	}
}

// Layout mengembalikan ukuran layar game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}
