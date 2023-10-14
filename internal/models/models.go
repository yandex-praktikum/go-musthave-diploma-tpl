package models

import (
	"github.com/golang-jwt/jwt/v4"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type LoyaltySystem struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual,omitempty"`
}

type SaveOrder struct {
	UserID   string
	OrderID  string
	Status   string
	Accrual  int
	Withdraw int
}
