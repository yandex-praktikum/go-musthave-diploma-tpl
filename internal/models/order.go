package models

import "time"

type Order struct {
	ID         uint   `gorm:"primary_key"`
	Number     string `gorm:"unique; not null"`
	UserID     uint   `gorm:"not null"`
	Status     string `gorm:"type:varchar(150) not null"`
	Accrual    int
	UploadedAt time.Time
}

type OrderDTO struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
