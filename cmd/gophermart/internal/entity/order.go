package entity

import (
	"time"
)

type Order struct {
	ID         string
	UserID     string
	Number     string
	UploadedAt time.Time
	Status     string
	Accrual    float32
}
