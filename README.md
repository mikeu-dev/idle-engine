# Idle Engine

Idle Engine adalah kerangka kerja (framework) dan engine modular untuk membuat game bergenre *idle/incremental* menggunakan bahasa pemrograman **Go** dan pustaka grafis **Ebitengine**.

## Fitur Utama

Proyek ini mengimplementasikan mekanik game idle klasik secara lengkap, terenskripsi, dinamis, dan thread-safe:
1. **Game Loop & Delta Time**: Perhitungan progress bar bisnis menggunakan delta time waktu nyata untuk kinerja yang konsisten di semua hardware.
2. **Konfigurasi Dinamis (YAML)**: Seluruh lini bisnis, upgrades, managers, dan achievements didefinisikan secara eksternal dalam [game_config.yaml](file:///c:/Users/dev/perusahaan/pst/workspaces/mikeu-dev/idle-engine/internal/infrastructure/config/game_config.yaml).
3. **Sistem NTP Anti-Cheat**: Memverifikasi kesesuaian waktu lokal dengan NTP Server tepercaya (`pool.ntp.org`) saat memuat data progres offline guna menangkal kecurangan pemalsuan jam sistem lokal.
4. **Mekanik Buy Max (Deret Geometri)**: Menghitung secara instan ($O(1)$) jumlah tingkat level maksimal yang bisa dibeli dengan sisa saldo saat ini menggunakan matematika geometri logaritma.
5. **GPS (Gain Per Second)**: Menghitung proyeksi pendapatan per detik untuk masing-masing bisnis serta GPS global pasif pemain.
6. **Floating Text Visual Juice**: Efek teks melayang interaktif berwarna hijau neon saat siklus produksi selesai untuk meningkatkan kepuasan bermain (*game feel*).
7. **Penyimpanan Terenkripsi (XOR)**: Menyimpan progres game secara terenkripsi menggunakan kunci XOR `0xAB` untuk menangkal manipulasi manual teks JSON polos pada berkas penyimpanan, lengkap dengan *legacy backward compatibility*.
8. **Prestige (Angel Investors)**: Mereset progres game demi mendapatkan Angel Investors yang memberikan peningkatan pendapatan permanen (+5% global per investor).

---

## Struktur Proyek

Proyek ini memisahkan logika domain secara bersih (DDD) dan memisahkan modul presentasi grafis:

- **`cmd/`**: Berisi entry point untuk aplikasi.
  - `cmd/sandbox/`: Lingkungan playground utama untuk menjalankan game sandbox.
- **`engine/`**: Berisi orchestrator utama permainan (`engine.go`) yang mengoordinasikan interaksi antar domain.
- **`internal/`**: Logika bisnis internal yang tidak diekspos ke luar proyek.
  - `internal/core/save/`: Manajemen enkripsi XOR & serialisasi data progres pemain (`SaveState`).
  - `internal/core/event/`: Mesin aturan pencapaian (*Achievements*) dan trigger kondisinya.
  - `internal/core/time/`: Manajemen waktu NTP anti-cheat.
  - `internal/domain/business/`: Lini bisnis, progress bar, formula pendapatan, dan perhitungan GPS.
  - `internal/domain/economy/`: Objek Wallet untuk mengatur balance transaksi.
  - `internal/domain/modifier/`: Struktur pengali statistik (Revenue & Speed).
  - `internal/domain/upgrade/`: Struktur kartu peningkatan bisnis.
  - `internal/domain/manager/`: Struktur manajer otomatisasi bisnis.
  - `internal/infrastructure/config/`: Pembaca berkas eksternal YAML untuk konfigurasi dinamis.
  - `internal/presentation/ebiten/`: Antarmuka visual 5 Tab UI berbasis **Ebitengine**.
- **`pkg/`**: Pustaka utilitas pembantu eksternal.
  - `pkg/mathutil/`: Perhitungan matematika logaritma deret geometri untuk Buy Max.

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
* **`[M]`** (Toggle Beli Maks): Mengubah mode pembelian upgrade bisnis antara **Beli 1x** dan **Beli Maks (MAX)**.
* **`[1]`**: Memulai produksi Lemonade Stand secara manual (jika belum otomatis).
* **`[Q]`**: Upgrade/Beli Lemonade Stand (Membeli level maks jika Mode Beli Maks aktif).
* **`[W]`**: Upgrade/Beli Newspaper Route.
* **`[E]`**: Upgrade/Beli Car Wash.

### 3. Aksi di Tab Upgrade (`[F2]`)
* **`[1]` / `[Q]`**: Membeli **Lemon Pitcher** (2x Pendapatan Lemonade Stand).
* **`[2]` / `[W]`**: Membeli **Newspaper Bag** (2x Kecepatan Newspaper Route).
* **`[3]` / `[E]`**: Membeli **Power Washer** (3x Pendapatan Car Wash).

### 4. Aksi di Tab Manajer (`[F3]`)
* **`[1]` / `[Q]`**: Mempekerjakan **Lemonade Manager** untuk mengotomatiskan produksi Lemonade Stand secara permanen.

### 5. Aksi di Tab Investor / Prestige (`[F5]`)
* **`[R]`** (Tekan Dua Kali): Tekan tombol `[R]` sebanyak **dua kali** dalam selang waktu 4 detik untuk mereset seluruh progres level bisnis, saldo, upgrade, dan manajer Anda demi mendapatkan **Angel Investors** baru.

### 6. Fitur Global
* **`[S]`**: Menyimpan progres game secara manual ke `savegame.json` (terenkripsi XOR).
* **`[L]`**: Memuat progres game secara manual dari `savegame.json` (mendukung decoding lama & baru).
* *Auto-save berjalan di background setiap 5 detik.*

