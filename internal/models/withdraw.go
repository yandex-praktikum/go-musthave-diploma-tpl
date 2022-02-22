package models

import "time"

type Withdrawl struct {
	ID           uint      `gorm:"primary_key" json:"-"`
	OrderID      uint      `gorm:"not null" json:"-"`
	Sum          float64   `json:"sum"`
	Processed_at time.Time `json:"processed_at"`
}

type Withdraw struct {
	Order        uint64    `json:"order" binding:"required"`
	Sum          float64   `json:"sum" binding:"required"`
	Processed_at time.Time `json:"processed_at,omitempty"`
}
