package models

import "time"

// User — модель пользователя (таблица users).
type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Active       bool
	CreatedAt    time.Time
}
