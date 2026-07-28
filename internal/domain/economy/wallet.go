package economy

import "sync"

// Wallet manages the game currency balance.
type Wallet struct {
	mu      sync.RWMutex
	balance float64
}

// NewWallet creates a new Wallet instance with an initial balance.
func NewWallet(initialBalance float64) *Wallet {
	return &Wallet{
		balance: initialBalance,
	}
}

// Balance returns the current wallet balance.
func (w *Wallet) Balance() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.balance
}

// Add adds a specified amount to the wallet.
func (w *Wallet) Add(amount float64) {
	if amount <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.balance += amount
}

// Spend deducts the amount from the wallet if balance is sufficient.
// Returns true if the transaction was successful, false otherwise.
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

// CanAfford checks if the wallet balance is sufficient for a specified amount.
func (w *Wallet) CanAfford(amount float64) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.balance >= amount
}
