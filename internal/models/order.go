package models

import "time"

// Order — заказ пользователя (таблица orders).
type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    *int // nil если начисления нет
	UploadedAt time.Time
}
