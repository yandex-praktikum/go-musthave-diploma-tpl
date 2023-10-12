package models

import "time"

type User struct {
	Id       int    `json:"-" db:"id"`
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required" db:"password_hash"`
	Salt     string `json:"salt"`
}

type Order struct {
	Number     string    `json:"number" db:"number"`
	Status     string    `json:"status" db:"status"`
	Accrual    float64   `json:"accrual" db:"sum"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaddate"`
}

type OrderResponse struct {
	Number  string  `json:"number" db:"number"`
	Status  string  `json:"status" db:"status"`
	Accrual float64 `json:"accrual" db:"accrual"`
}

type Balance struct {
	Current   float64 `json:"current" db:"current"`
	Withdrawn float64 `json:"withdrawn" db:"withdrawn"`
}

type Withdraw struct {
	Order string  `json:"order" db:"number"`
	Sum   float64 `json:"sum" db:"sum"`
}

type WithdrawResponse struct {
	Order     string    `json:"order" db:"number"`
	Sum       float64   `json:"sum" db:"sum"`
	Processed time.Time `json:"processed_at" db:"processed"`
}
