package repository

import (
	"fmt"
)

//add to cash ============================================================
func (r *Repository) AddToCash(key string, value string) {
	r.cash.Store(key, value)
}

//get from cash ==========================================================
func (r *Repository) GetFromCash(key string) (string, bool) {
	value, ok := r.cash.Load(key)
	return fmt.Sprint(value), ok
}

//remove from cash =======================================================
func (r *Repository) RemoveFromCash(key string) {
	r.cash.Delete(key)
}

//print cash =============================================================
func (r *Repository) PrintCash() string {
	return fmt.Sprintf("cash: %v", r.cash)
}
