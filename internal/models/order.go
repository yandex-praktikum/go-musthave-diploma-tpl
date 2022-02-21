package models

import "time"

type Order struct {
	ID         uint   `gorm:"primary_key"`
	Number     string `gorm:"type:varchar(150) unique not null"`
	UserID     uint   `gorm:"not null"`
	Status     string `gorm:"type:varchar(150) not null"`
	Accrual    float64
	UploadedAt time.Time
}
