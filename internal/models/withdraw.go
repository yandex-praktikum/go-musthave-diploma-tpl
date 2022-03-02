package models

import "time"

type Withdrawal struct {
	ID          uint `gorm:"primary_key"`
	OrderID     uint `gorm:"not null"`
	Sum         float64
	ProcessedAt time.Time
}

type Withdraw struct {
	Order       string    `json:"order" binding:"required"`
	Sum         float64   `json:"sum" binding:"required"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
