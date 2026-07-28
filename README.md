# Idle Engine

Idle Engine adalah kerangka kerja (framework) dan engine modular untuk membuat game bergenre *idle/incremental* menggunakan bahasa pemrograman **Go** dan pustaka grafis **Ebitengine**.

## Fitur Utama

Proyek ini mengimplementasikan mekanik game idle klasik secara lengkap dan thread-safe:
1. **Game Loop & Delta Time**: Perhitungan progress bar bisnis menggunakan delta time waktu nyata untuk kinerja yang konsisten di semua hardware.
2. **Otomatisasi Bisnis**: Mempekerjakan manajer untuk mengotomatiskan lini bisnis (seperti Lemonade Stand) secara dinamis.
3. **Peningkatan Statistik (Upgrades & Modifiers)**: Sistem peningkatan multiplier pendapatan dan kecepatan per lini bisnis.
4. **Sistem Simpan/Muat (Save/Load)**: Menyimpan progres game secara otomatis (setiap 5 detik) atau manual ke dalam format JSON lokal (`savegame.json`).
5. **Akumulasi Offline (Offline Progress)**: Menghitung secara efisien pendapatan ketika game ditutup (maksimal 12 jam offline) menggunakan pembagian modulo bebas lag ($O(1)$).
6. **Sistem Pencapaian (Achievements)**: Membuka pencapaian tertentu untuk mendapatkan bonus pengali pendapatan global secara permanen.
7. **Sistem Prestige (Angel Investors)**: Mereset progres game demi mendapatkan Angel Investors yang memberikan peningkatan pendapatan permanen (+5% global per investor).

---

## Struktur Proyek

Proyek ini memisahkan logika domain secara bersih (DDD) dan memisahkan modul presentasi grafis:

- **`cmd/`**: Berisi entry point untuk aplikasi.
  - `cmd/sandbox/`: Lingkungan playground utama untuk menjalankan game sandbox.
- **`engine/`**: Berisi orchestrator utama permainan (`engine.go`) yang mengoordinasikan interaksi antar domain.
- **`internal/`**: Logika bisnis internal yang tidak diekspos ke luar proyek.
  - `internal/core/save/`: Manajemen serialisasi JSON untuk data progres pemain (`SaveState`).
  - `internal/core/event/`: Mesin aturan pencapaian (*Achievements*) dan trigger kondisinya.
  - `internal/domain/business/`: Definisi entitas lini bisnis, progress bar, dan formula biaya upgrade level.
  - `internal/domain/economy/`: Objek Wallet untuk mengatur balance transaksi.
  - `internal/domain/modifier/`: Struktur pengali statistik (Revenue & Speed).
  - `internal/domain/upgrade/`: Struktur kartu peningkatan bisnis.
  - `internal/domain/manager/`: Struktur manajer otomatisasi bisnis.
  - `internal/presentation/ebiten/`: Antarmuka visual 5 Tab UI berbasis **Ebitengine**.

---

## Cara Menjalankan Aplikasi

Jalankan perintah berikut di direktori aktif proyek Anda untuk memulai sandbox Ebitengine:

```powershell
go run ./cmd/sandbox/main.go
```

Untuk menjalankan seluruh unit test terotomatisasi:

```powershell
go test ./...
```

---

## Kontrol Permainan (Sandbox UI)

### 1. Navigasi Tab Halaman
Anda dapat beralih halaman tab menggunakan tombol **`[Tab]`** pada keyboard secara bergantian, atau berpindah secara instan menggunakan tombol berikut:
* **`[F1]`**: Tab Lini Bisnis
* **`[F2]`**: Tab Peningkatan (Upgrade)
* **`[F3]`**: Tab Manajer (Automation)
* **`[F4]`**: Tab Pencapaian (Achievements)
* **`[F5]`**: Tab Investor (Prestige)

### 2. Aksi di Tab Bisnis (`[F1]`)
* **`[1]`**: Memulai produksi Lemonade Stand (jika belum otomatis).
* **`[Q]`**: Meningkatkan level Lemonade Stand (Biaya poin).
* **`[W]`**: Meningkatkan level Newspaper Route.
* **`[E]`**: Meningkatkan level Car Wash.

### 3. Aksi di Tab Upgrade (`[F2]`)
* **`[1]` / `[Q]`**: Membeli **Lemon Pitcher** (2x Pendapatan Lemonade Stand).
* **`[2]` / `[W]`**: Membeli **Newspaper Bag** (2x Kecepatan Newspaper Route).
* **`[3]` / `[E]`**: Membeli **Power Washer** (3x Pendapatan Car Wash).

### 4. Aksi di Tab Manajer (`[F3]`)
* **`[1]` / `[Q]`**: Mempekerjakan **Lemonade Manager** untuk mengotomatiskan produksi Lemonade Stand secara permanen.

### 5. Aksi di Tab Investor / Prestige (`[F5]`)
* **`[R]`** (Tolak Dua Kali): Tekan tombol `[R]` sebanyak **dua kali** dalam selang waktu 4 detik untuk mereset seluruh progres level bisnis, saldo, upgrade, dan manajer Anda demi mendapatkan **Angel Investors** baru.

### 6. Fitur Global
* **`[S]`**: Menyimpan progres game secara manual ke `savegame.json`.
* **`[L]`**: Memuat progres game secara manual dari `savegame.json`.
* *Auto-save berjalan di background setiap 5 detik.*
