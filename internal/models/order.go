package models

import "time"

type Order struct {
	ID         uint      `gorm:"primary_key" json:"-"`
	Number     uint64    `gorm:"unique; not null" json:"number"`
	UserID     uint      `gorm:"not null" json:"-"`
	Status     string    `gorm:"type:varchar(150) not null" json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
