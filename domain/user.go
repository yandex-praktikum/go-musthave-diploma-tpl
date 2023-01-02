package domain

import "time"

type User struct {
	Login    string `json:"login"    db:"login"`
	Password string `json:"password" db:"password"`
	Hash     string `json:"hash"     db:"hash"`
	Token    string `json:"token"    db:"token"`
	ID       int64  `json:"id"       db:"id"`
	Name     string
}

func (u *User) IsValid() bool {
	if u.Login == "" {
		return false
	}
	if u.Password == "" {
		return false
	}
	return true
}

type Balance struct {
	// Current   decimal.Decimal `json:"current"`
	// Withdrawn decimal.Decimal `json:"withdrawn"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	Order       string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at"`
}
