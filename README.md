# Idle Engine

Idle Engine is a modular framework and engine for building idle/incremental games using the **Go** programming language and the **Ebitengine** graphics library.

## Key Features

The project implements a complete, thread-safe set of classic idle game mechanics:
1. **Game Loop & Delta Time**: Business progress bar calculations use real-time delta time for consistent performance across all hardware.
2. **Dynamic Configuration (YAML)**: All business lines, upgrades, managers, and achievements are defined externally in [game_config.yaml](file:///c:/Users/dev/perusahaan/pst/workspaces/mikeu-dev/idle-engine/internal/infrastructure/config/game_config.yaml).
3. **NTP Anti-Cheat System**: Verifies system time against a trusted NTP server (`pool.ntp.org`) during offline progress loads to prevent local time manipulation cheats.
4. **Buy Max Mechanic (Geometric Series)**: Instantly calculates ($O(1)$ complexity) the maximum affordable level upgrades with current wallet balances using closed-form logarithmic geometric series.
5. **GPS (Gain Per Second)**: Dynamically calculates the gain per second contribution for each individual business and the global passive GPS of the player.
6. **Floating Text Visual Juice**: Interactive neon green floating text notifications appear when a production cycle completes to enhance player feedback.
7. **Encrypted Saves (XOR)**: Progres is saved securely using a XOR obfuscation algorithm with a secret key (`0xAB`) to prevent manual JSON save file editing, with *legacy backward compatibility* for plain JSON files.
8. **Prestige System (Angel Investors)**: Reset game progress to claim Angel Investors that grant permanent global income boosts (+5% global multiplier per angel).

---

## Project Structure

This project enforces clean domain separation (DDD) and separates visual presentation logic:

- **`cmd/`**: Entry points for the application.
  - `cmd/sandbox/`: The main playground environment to run the sandbox game.
- **`engine/`**: The orchestrator (`engine.go`) coordinating interactions between domains.
- **`internal/`**: Core logic not exposed outside the project.
  - `internal/core/save/`: XOR encryption and player progress serialisation (`SaveState`).
  - `internal/core/event/`: Rule engine for achievements and trigger conditions.
  - `internal/core/time/`: NTP anti-cheat and clock validation.
  - `internal/domain/business/`: Business entities, progress bars, income formula, and GPS calculations.
  - `internal/domain/economy/`: Wallet object managing balance transactions.
  - `internal/domain/modifier/`: Multipliers for speed and revenue.
  - `internal/domain/upgrade/`: Business upgrade cards.
  - `internal/domain/manager/`: Manager automation cards.
  - `internal/infrastructure/config/`: External YAML configuration loader.
  - `internal/presentation/ebiten/`: Visual user interface with a 5-tab menu built on **Ebitengine**.
- **`pkg/`**: External helper utilities.
  - `pkg/mathutil/`: Logarithmic geometric series formulas for Buy Max.

---

## Getting Started

Run the following command in the project root directory to start the Ebitengine sandbox:

```powershell
go run ./cmd/sandbox/main.go
```

To run all automated unit tests:

```powershell
go test ./...
```

---

## Controls & Sandbox UI

### 1. Tab Menu Navigation
You can cycle through pages using the **`[Tab]`** key, or switch directly to a tab using:
* **`[F1]`**: Business Lines Tab
* **`[F2]`**: Upgrades Tab
* **`[F3]`**: Managers Tab
* **`[F4]`**: Achievements Tab
* **`[F5]`**: Investors (Prestige) Tab

### 2. Business Lines Tab (`[F1]`)
* **`[M]`** (Toggle Buy Mode): Toggle upgrade purchase mode between **Buy 1x** and **Buy Max (MAX)**.
* **`[1]`**: Start Lemonade Stand production manually (if not automated).
* **`[Q]`**: Upgrade/Buy Lemonade Stand.
* **`[W]`**: Upgrade/Buy Newspaper Route.
* **`[E]`**: Upgrade/Buy Car Wash.

### 3. Upgrades Tab (`[F2]`)
* **`[1]` / `[Q]`**: Buy **Lemon Pitcher** (2x Lemonade Stand revenue).
* **`[2]` / `[W]`**: Buy **Newspaper Bag** (2x Newspaper Route speed).
* **`[3]` / `[E]`**: Buy **Power Washer** (3x Car Wash revenue).
* **`[U]`** (Super Boost): Buy Super Boost (2x global speed for 30s, costs 50 points).
* **`[I]`** (Time Warp): Buy Time Warp (Instant +1 hour of passive automated income, costs 150 points).

### 4. Managers Tab (`[F3]`)
* **`[1]` / `[Q]`**: Hire **Lemonade Manager** to automate Lemonade Stand production permanently.

### 5. Investors / Prestige Tab (`[F5]`)
* **`[R]`** (Press Twice): Press `[R]` twice within 4 seconds to reset progress (balance, levels, upgrades, managers) and claim **Angel Investors**.

### 6. Global Features
* **Random Events**: Economic breaking news events trigger automatically every 45 seconds to modify business performance dynamically.
* **`[T]`**: Toggle UI color themes (Catppuccin Mocha, Cyberpunk Neon, Nordic Frost) instantly.
* **`[S]`**: Save progress manually to `savegame.json` (XOR encrypted, includes active boost timers).
* **`[L]`**: Load progress manually from `savegame.json` (supports both legacy plain JSON and encrypted formats).
* *Auto-save runs in the background every 5 seconds.*

