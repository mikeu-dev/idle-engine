package economy

import "sync"

// Wallet mengelola saldo mata uang dalam game.
type Wallet struct {
	mu      sync.RWMutex
	balance float64
}

// NewWallet membuat instansi Wallet baru dengan saldo awal.
func NewWallet(initialBalance float64) *Wallet {
	return &Wallet{
		balance: initialBalance,
	}
}

// Balance mengembalikan saldo dompet saat ini.
func (w *Wallet) Balance() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.balance
}

// Add menambahkan sejumlah uang ke dalam dompet.
func (w *Wallet) Add(amount float64) {
	if amount <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.balance += amount
}

// Spend membelanjakan uang dari dompet jika saldo mencukupi.
// Mengembalikan true jika transaksi sukses, dan false jika saldo kurang.
func (w *Wallet) Spend(amount float64) bool {
	if amount <= 0 {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.balance >= amount {
		w.balance -= amount
		return true
	}
	return false
}

// CanAfford memeriksa apakah saldo dompet mencukupi untuk nominal tertentu.
func (w *Wallet) CanAfford(amount float64) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.balance >= amount
}
