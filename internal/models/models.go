package models

import "time"

type User struct {
	Id       int    `json:"-" db:"id"`
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required" db:"password_hash"`
	Salt     string `json:"salt"`
}

type Order struct {
	Number      string    `json:"number" db:"number"`
	Status      string    `json:"status" db:"status"`
	Accrual     float64   `json:"accrual" db:"sum"`
	Uploaded_at time.Time `json:"uploaded_at" db:"uploaddate"`
}
