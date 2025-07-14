package models

import "time"

type User struct {
	ID           int64     `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type Order struct {
	ID          int64     `db:"id"`
	OrderNumber string    `db:"order_number"`
	UserID      int64     `db:"user_id"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
}
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
