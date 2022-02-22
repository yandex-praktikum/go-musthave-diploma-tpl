package models

import "time"

type Withdrawal struct {
	ID           uint `gorm:"primary_key"`
	OrderID      uint `gorm:"not null"`
	Sum          float64
	Processed_at time.Time
}

type Withdraw struct {
	Order        string    `json:"order" binding:"required"`
	Sum          float64   `json:"sum" binding:"required"`
	Processed_at time.Time `json:"processed_at,omitempty"`
}
