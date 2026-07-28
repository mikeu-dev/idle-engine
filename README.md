# Idle Engine

Idle Engine adalah kerangka kerja (framework) dan engine sederhana untuk membuat game bergenre *idle/incremental* menggunakan bahasa pemrograman **Go** dan pustaka grafis **Ebitengine**.

## Struktur Proyek

Proyek ini menggunakan tata letak standar Go:

- **`cmd/`**: Berisi entry point untuk aplikasi.
  - `cmd/sandbox/`: Lingkungan percobaan (playground) untuk menguji fitur engine secara langsung.
  - `cmd/sample-game/`: Contoh implementasi game utuh menggunakan engine ini.
- **`engine/`**: Logika inti permainan (state engine, game loop, kalkulasi delta waktu).
- **`internal/`**: Logika bisnis internal yang tidak diekspos ke luar proyek.
  - `internal/core/`: Manajemen tick, sistem event, dan save/load game.
  - `internal/domain/`: Entitas game seperti *business*, *economy*, *upgrades*, *modifiers*, dan *production*.
  - `internal/infrastructure/`: Konfigurasi sistem.
  - `internal/presentation/`: Handler UI/rendering (saat ini menggunakan Ebitengine).
- **`pkg/`**: Kode utilitas yang dapat digunakan kembali oleh proyek luar (jika ada).

## Prasyarat

Sebelum menjalankan proyek ini, pastikan Anda telah menginstal:
- **Go** (versi 1.22 ke atas direkomendasikan, proyek ini diuji menggunakan Go 1.26)

## Cara Menjalankan

Untuk menjalankan lingkungan sandbox:

```bash
go run ./cmd/sandbox/main.go
```

## Kontrol Permainan (Sandbox)
- **Tekan tombol `[U]`** di keyboard untuk membeli Upgrade (membutuhkan 10 poin) guna meningkatkan produksi poin per detik.
