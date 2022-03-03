package models

import "time"

type Withdrawal struct {
	ID          uint `gorm:"primary_key"`
	OrderID     uint `gorm:"not null"`
	Sum         int
	ProcessedAt time.Time
}

type WithdrawalDTO struct {
	Order       string    `json:"order" binding:"required"`
	Sum         float64   `json:"sum" binding:"required"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
