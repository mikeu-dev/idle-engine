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

// FloatingText represents a small floating text on the screen.
type FloatingText struct {
	x, y      float32
	text      string
	alpha     float32
	createdAt time.Time
}

// UITheme defines an integrated color palette for the game interface.
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

// Game implements the ebiten.Game interface.
type Game struct {
	engine                *engine.Engine
	lastAutoSave          time.Time
	notification          string
	notificationExpiry    time.Time
	activeTab             int // 0: Business, 1: Upgrade, 2: Manager, 3: Achievements, 4: Investor
	prestigeConfirm       bool
	prestigeConfirmExpiry time.Time
	buyMaxMode            bool
	prevProgress          []time.Duration
	floatingTexts         []FloatingText
	activeThemeIndex      int
	themes                []UITheme
}

// NewGame creates a new Game instance with the given engine.
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

// SetOfflineNotification sets the offline revenue notification message when the game is first opened.
func (g *Game) SetOfflineNotification(revenue float64) {
	g.notification = fmt.Sprintf("[WELCOME BACK! OFFLINE REVENUE: +%.2f POINTS]", revenue)
	g.notificationExpiry = time.Now().Add(5 * time.Second)
}

// Update processes input and updates the engine and the save-load cycle.
func (g *Game) Update() error {
	g.engine.Update()

	// Reset prestige confirm state if expired
	if g.prestigeConfirm && time.Now().After(g.prestigeConfirmExpiry) {
		g.prestigeConfirm = false
	}

	// Toggle Buy Mode with M (Global)
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.buyMaxMode = !g.buyMaxMode
		if g.buyMaxMode {
			g.notification = "[UPGRADE MODE: BUY MAX (MAX)]"
		} else {
			g.notification = "[UPGRADE MODE: BUY 1x]"
		}
		g.notificationExpiry = time.Now().Add(2 * time.Second)
	}

	// Toggle UI Color Theme with T (Global)
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.activeThemeIndex = (g.activeThemeIndex + 1) % len(g.themes)
		g.notification = fmt.Sprintf("[UI COLOR THEME ACTIVE: %s]", g.themes[g.activeThemeIndex].Name)
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

	// Detect Business Cycle Completion for Floating Text
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
			g.spawnFloatingText(550, yPos, fmt.Sprintf("+%.2f Points", b.Income()))
		}
		g.prevProgress[i] = currProg
	}

	// Cycle tabs with Tab key
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.activeTab = (g.activeTab + 1) % 5
		g.showTabChangeNotification()
	}

	// Switch directly to tabs with F1-F5
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

	// Business Tab Hotkeys (activeTab == 0)
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

	// Upgrade Tab Hotkeys (activeTab == 1)
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
				g.notification = fmt.Sprintf("[%s ALREADY PURCHASED!]", upg.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else if g.engine.BuyUpgradeCard(boughtIdx) {
				g.notification = fmt.Sprintf("[%s PURCHASED SUCCESSFULLY!]", upg.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else {
				g.notification = fmt.Sprintf("[FAILED TO BUY %s (INSUFFICIENT BALANCE)]", upg.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			}
		}

		// Temporary Booster triggers
		if inpututil.IsKeyJustPressed(ebiten.KeyU) {
			if g.engine.TriggerSuperBoost() {
				g.notification = "[SUPER BOOST ACTIVATED: 2X SPEED!]"
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else {
				g.notification = "[FAILED TO ACTIVATE SUPER BOOST (REQUIRES 50 POINTS)]"
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyI) {
			if g.engine.TriggerTimeWarp() {
				g.notification = "[TIME WARP SUCCESS: INSTANT +1 HOUR OF PASSIVE INCOME!]"
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
					g.notification = "[FAILED: TIME WARP REQUIRES AT LEAST 1 AUTOMATED BUSINESS]"
				} else {
					g.notification = "[FAILED TIME WARP (REQUIRES 150 POINTS)]"
				}
				g.notificationExpiry = time.Now().Add(3 * time.Second)
			}
		}
	}

	// Manager Tab Hotkeys (activeTab == 2)
	if g.activeTab == 2 {
		hiredIdx := -1
		if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad1) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			hiredIdx = 0
		}

		if hiredIdx != -1 {
			m := g.engine.GetManagers()[hiredIdx]
			if m.IsHired {
				g.notification = fmt.Sprintf("[%s ALREADY HIRED!]", m.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else if g.engine.BuyManager(hiredIdx) {
				g.notification = fmt.Sprintf("[%s HIRED SUCCESSFULLY!]", m.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			} else {
				g.notification = fmt.Sprintf("[FAILED TO HIRE %s (INSUFFICIENT BALANCE)]", m.Name)
				g.notificationExpiry = time.Now().Add(2 * time.Second)
			}
		}
	}

	// Investor / Prestige Tab Hotkeys (activeTab == 4)
	if g.activeTab == 4 {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			claimable := g.engine.CalculateAngelsToClaim()
			if claimable <= 0 {
				g.notification = "[CANNOT PRESTIGE (NO INVESTORS TO CLAIM)]"
				g.notificationExpiry = time.Now().Add(3 * time.Second)
				g.prestigeConfirm = false
			} else if g.prestigeConfirm && time.Now().Before(g.prestigeConfirmExpiry) {
				// Run reset and claim
				if g.engine.ClaimPrestige() {
					g.notification = "[PRESTIGE SUCCESS! PROGRESS RESET WITH NEW INVESTOR BONUS]"
					g.notificationExpiry = time.Now().Add(4 * time.Second)
				}
				g.prestigeConfirm = false
			} else {
				// Enter confirmation state
				g.prestigeConfirm = true
				g.prestigeConfirmExpiry = time.Now().Add(4 * time.Second)
				g.notification = "[PRESS 'R' AGAIN TO CONFIRM RESET & CLAIM!]"
				g.notificationExpiry = time.Now().Add(4 * time.Second)
			}
		}
	}

	// Manual Save/Load Hotkeys (Global)
	saveFile := "savegame.json"
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		state := g.engine.ExportState()
		if err := save.SaveToFile(saveFile, state); err != nil {
			g.notification = "[FAILED TO SAVE GAME!]"
		} else {
			g.notification = "[GAME SAVED SUCCESSFULLY!]"
		}
		g.notificationExpiry = time.Now().Add(3 * time.Second)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		state, err := save.LoadFromFile(saveFile)
		if err != nil {
			g.notification = "[FAILED TO LOAD GAME!]"
		} else {
			g.engine.ImportState(state)
			g.notification = "[GAME LOADED SUCCESSFULLY!]"
		}
		g.notificationExpiry = time.Now().Add(3 * time.Second)
	}

	// Auto-Save every 5 seconds
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
		0: "BUSINESS LINES",
		1: "UPGRADES",
		2: "MANAGERS (AUTOMATION)",
		3: "ACHIEVEMENTS",
		4: "ANGEL INVESTORS (PRESTIGE)",
	}[g.activeTab]
	g.notification = fmt.Sprintf("[ACTIVE TAB: %s]", tabName)
	g.notificationExpiry = time.Now().Add(1500 * time.Millisecond)
}

// Draw renders the game interface to the screen.
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
		gpsStr = fmt.Sprintf("BALANCE: %.2f POINTS (+%.2f/sec) [BOOST: %.1fs]", balance, g.engine.GetTotalGPS(), boostDur.Seconds())
	} else {
		gpsStr = fmt.Sprintf("BALANCE: %.2f POINTS (+%.2f/sec)", balance, g.engine.GetTotalGPS())
	}
	ebitenutil.DebugPrintAt(screen, gpsStr, 20, 30)

	modeStr := "BUY MODE: [1x] (Press [M] for Max)"
	if g.buyMaxMode {
		modeStr = "BUY MODE: [MAX] (Press [M] for 1x)"
	}
	ebitenutil.DebugPrintAt(screen, modeStr, 340, 30)

	// Draw active notification if not expired
	if time.Now().Before(g.notificationExpiry) && g.notification != "" {
		ebitenutil.DebugPrintAt(screen, g.notification, 320, 12)
	}

	// Draw Tab Header (5 Tab UI) with dynamic colors according to active theme
	tabs := []struct {
		tabIdx int
		title  string
		hotkey string
		color  color.RGBA
	}{
		{0, "BUSINESS", "F1", theme.Blue},
		{1, "UPGRADES", "F2", theme.Green},
		{2, "MANAGERS", "F3", theme.Yellow},
		{3, "ACHIEVE", "F4", theme.Pink},
		{4, "PRESTIGE", "F5", theme.Orange},
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

	ebitenutil.DebugPrintAt(screen, "(Press F1-F5 or [TAB])", 475, 53)

	// TAB 1: BUSINESS
	if g.activeTab == 0 {
		businesses := g.engine.GetBusinesses()
		for i, b := range businesses {
			y := 88 + i*112

			// Business Title & Level with Speed Milestone Targets
			var statusText string
			if b.IsOwned() {
				statusText = fmt.Sprintf("Lv. %d/%d (Speed 2x)", b.GetLevel(), b.GetNextMilestone())
			} else {
				statusText = "Locked"
			}
			
			businessTitle := fmt.Sprintf("%d. %s (%s)", i+1, b.Name, statusText)
			ebitenutil.DebugPrintAt(screen, businessTitle, 20, y)

			// Production Details & Upgrade Cost
			var detailText string
			var actionText string

			cost := b.Cost()
			canAfford := wallet.CanAfford(cost)
			
			var costColor string
			if canAfford {
				costColor = "Buy"
			} else {
				costColor = "Locked"
			}

			// Shortcut key configurations
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
				detailText = fmt.Sprintf("Revenue: %.2f Points / %s (+%.2f/sec)", b.Income(), b.Duration, b.GetGPS())
			} else {
				detailText = fmt.Sprintf("Base Revenue: %.2f Points / %s", b.BaseIncome, b.Duration)
			}

			if g.buyMaxMode {
				levels, totalCost := mathutil.CalculateMaxLevelsAffordable(b.BaseCost, b.CostMultiplier, b.GetLevel(), balance)
				if levels > 0 {
					actionText = fmt.Sprintf("Buy Max: Press [%s] (Get +%d Levels, Cost: %.2f) [Buy]", keyName, levels, totalCost)
				} else {
					actionText = fmt.Sprintf("Buy Max: Press [%s] (Cost: %.2f) [Locked]", keyName, cost)
				}
			} else {
				if b.IsAutomated {
					actionText = fmt.Sprintf("Upgrade: Press [%s] (Cost: %.2f) [%s] (Auto)", keyName, cost, costColor)
				} else {
					actionText = fmt.Sprintf("Upgrade: Press [%s] (Cost: %.2f) [%s] | Start: Press [%d]", keyName, cost, costColor, i+1)
				}
			}

			ebitenutil.DebugPrintAt(screen, detailText, 20, y+20)
			ebitenutil.DebugPrintAt(screen, actionText, 20, y+38)

			// Progress Bar
			barX := 20
			barY := y + 58
			barW := 600
			barH := 12

			// Background Progress Bar from active theme
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

		ebitenutil.DebugPrintAt(screen, "Hint: Automated businesses will start producing on their own once purchased.", 20, 422)
	}

	// TAB 2: UPGRADE
	if g.activeTab == 1 {
		// 1. Upgrade Cards on the Left Side (x=20 to x=400)
		upgrades := g.engine.GetUpgrades()
		for i, upg := range upgrades {
			y := 88 + i*108

			// Draw card background (box lebar 380) from active theme
			vector.DrawFilledRect(screen, 20, float32(y), 380, 95, theme.CardBg, false)

			titleText := fmt.Sprintf("[%d] %s", i+1, upg.Name)
			if upg.IsPurchased {
				titleText += " (PURCHASED)"
			}
			ebitenutil.DebugPrintAt(screen, titleText, 35, y+10)
			ebitenutil.DebugPrintAt(screen, upg.Description, 35, y+30)

			var actionText string
			if upg.IsPurchased {
				actionText = "Stat: Upgrade Active"
			} else {
				canAfford := wallet.CanAfford(upg.Cost)
				statusColor := "Locked"
				if canAfford {
					statusColor = "Buy"
				}
				actionText = fmt.Sprintf("Buy: Press [%s] (Cost: %.2f) [%s]", map[int]string{0: "Q", 1: "W", 2: "E"}[i], upg.Cost, statusColor)
			}
			ebitenutil.DebugPrintAt(screen, actionText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Hint: Upgrades modify global permanent multipliers.", 20, 422)

		// 2. Temporary Booster Shop on the Right Side (x=420 to x=620)
		ebitenutil.DebugPrintAt(screen, "=== TEMPORARY BOOSTER SHOP ===", 420, 88)

		// Kartu Super Boost (y=110)
		vector.DrawFilledRect(screen, 420, 110, 200, 100, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "[U] SUPER BOOST", 435, 120)
		ebitenutil.DebugPrintAt(screen, "2x Speed (30s)", 435, 140)
		ebitenutil.DebugPrintAt(screen, "Cost: 50.00 Points", 435, 160)
		boostStatus := "[Locked]"
		if wallet.CanAfford(50.0) {
			boostStatus = "[Buy]"
		}
		ebitenutil.DebugPrintAt(screen, boostStatus, 435, 180)

		// Kartu Time Warp (y=230)
		vector.DrawFilledRect(screen, 420, 230, 200, 100, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "[I] TIME WARP", 435, 240)
		ebitenutil.DebugPrintAt(screen, "Instant +1 Hr Auto Revenue", 435, 260)
		ebitenutil.DebugPrintAt(screen, "Cost: 150.00 Points", 435, 280)
		warpStatus := "[Locked]"
		if wallet.CanAfford(150.0) {
			warpStatus = "[Buy]"
		}
		ebitenutil.DebugPrintAt(screen, warpStatus, 435, 300)
	}

	// TAB 3: MANAGERS
	if g.activeTab == 2 {
		managers := g.engine.GetManagers()
		for i, m := range managers {
			y := 88 + i*108

			// Draw card background (box) from active theme
			vector.DrawFilledRect(screen, 20, float32(y), 600, 95, theme.CardBg, false)

			titleText := fmt.Sprintf("[%d] %s", i+1, m.Name)
			if m.IsHired {
				titleText += " (HIRED)"
			}
			ebitenutil.DebugPrintAt(screen, titleText, 35, y+10)
			ebitenutil.DebugPrintAt(screen, m.Description, 35, y+30)

			var actionText string
			if m.IsHired {
				actionText = "Automation: Active"
			} else {
				canAfford := wallet.CanAfford(m.Cost)
				statusColor := "Locked"
				if canAfford {
					statusColor = "Buy"
				}
				actionText = fmt.Sprintf("Hire: Press [%d] / [%s] (Cost: %.2f) [%s]", i+1, map[int]string{0: "Q", 1: "W", 2: "E"}[i], m.Cost, statusColor)
			}
			ebitenutil.DebugPrintAt(screen, actionText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Hint: Managers automate manual Business Lines to run on their own.", 20, 422)
	}

	// TAB 4: ACHIEVEMENTS
	if g.activeTab == 3 {
		achievements := g.engine.GetAchievements()
		for i, ach := range achievements {
			y := 88 + i*108

			// Draw card background (box) from active theme
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
				titleText += " (UNLOCKED)"
			} else {
				titleText += " (LOCKED)"
			}
			ebitenutil.DebugPrintAt(screen, titleText, 35, y+10)
			ebitenutil.DebugPrintAt(screen, ach.Description, 35, y+30)

			var statusText string
			if ach.IsUnlocked {
				statusText = "Status: +10%/20% Modifier Active"
			} else {
				statusText = "Status: Conditions not met"
			}
			ebitenutil.DebugPrintAt(screen, statusText, 35, y+55)
		}

		ebitenutil.DebugPrintAt(screen, "Hint: Achievements unlock automatically when conditions are met, giving permanent global revenue boosts.", 20, 422)
	}

	// TAB 5: INVESTOR / PRESTIGE
	if g.activeTab == 4 {
		lifetime := g.engine.GetLifetimeEarnings()
		angels := g.engine.GetAngels()
		claimable := g.engine.CalculateAngelsToClaim()

		// 1. Statistics Card from active theme
		vector.DrawFilledRect(screen, 20, 88, 600, 110, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "LIFETIME STATISTICS:", 35, 98)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Total Accumulated Revenue : %.2f Points", lifetime), 35, 120)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Angel Investors Owned     : %d Angels", angels), 35, 142)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Global Revenue Multiplier : +%d%% Revenue (Global)", angels*5), 35, 164)

		// 2. Prestige Card from active theme
		vector.DrawFilledRect(screen, 20, 218, 600, 130, theme.CardBg, false)
		ebitenutil.DebugPrintAt(screen, "RESET PRESTIGE (CLAIM ANGEL INVESTORS):", 35, 228)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - New Angels to Claim      : +%d Angels", claimable), 35, 250)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" - Expected Global Multiplier: +%d%% Revenue (Global)", (angels+claimable)*5), 35, 272)

		var prestigeText string
		if claimable <= 0 {
			prestigeText = "Action: No new angels to claim (Requires at least 100 Lifetime Points)."
		} else if g.prestigeConfirm {
			prestigeText = "[PRESS 'R' AGAIN TO CONFIRM RESET & CLAIM!]"
		} else {
			prestigeText = "Action: Press [R] to trigger prestige reset & claim new angels."
		}
		ebitenutil.DebugPrintAt(screen, prestigeText, 35, 305)

		// Keterangan
		ebitenutil.DebugPrintAt(screen, "Info: Performing Prestige resets balance, business levels, upgrades, and managers.", 20, 365)
		ebitenutil.DebugPrintAt(screen, "However, Angel Investors grant +5% permanent global revenue boost. Achievements are NOT reset.", 20, 385)
	}

	// Footer (Global) & Active Random Event Ticker
	activeEvt, eventDur := g.engine.GetActiveEvent()
	if activeEvt != nil {
		eventStr := fmt.Sprintf("[BREAKING NEWS] %s: %s (Remaining %.1fs)", activeEvt.Title, activeEvt.Description, eventDur.Seconds())
		ebitenutil.DebugPrintAt(screen, eventStr, 20, 435)
	}
	ebitenutil.DebugPrintAt(screen, "Features: [S] Manual Save | [L] Manual Load | Auto-save active (5s)", 20, 455)

	// Render Floating Texts
	for _, ft := range g.floatingTexts {
		ebitenutil.DebugPrintAt(screen, ft.text, int(ft.x), int(ft.y))
	}
}

// Layout returns the virtual screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}
