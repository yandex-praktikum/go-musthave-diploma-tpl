package models

import "time"

type Users struct {
	ID       int32   `gorm:"serial primary_key"`
	Login    string  `gorm:"not null"`
	Password string  `gorm:"not null"`
	Balance  float32 `gorm:"default 0"`
}

// type Orders struct {
// 	ID            int32     `gorm:"serial primary_key"`
// 	Number        string    `gorm:"text not null"`
// 	UserLogin     string    `gorm:"text not null"`
// 	UploadedAt    time.Time `gorm:"timestamp with time zone default now() not null"`
// 	OperationType string    `gorm:"text default 'accrual'"`
// 	Status        string    `gorm:"not null default 'new'"`
// 	Amount        float32   `gorm:"default 0 not null"`
// }

type Order struct {
	ID         int32     `gorm:"primaryKey" json:"-"`
	Number     string    `gorm:"not null;unique" json:"number"`
	UserLogin  string    `json:"-"`
	UploadedAt time.Time `json:"uploaded_at"`
	Accrual    float32   `json:"accrual"`
	Status     string    `json:"status"`
}

type User struct {
	Login    string `gorm:"column:login" json:"login"`
	Password string `gorm:"column:password" json:"password"`
	Balance  int    `gorm:"column:balance" json:"balance"`
}

// ID        uint      `gorm:"primaryKey" json:"-"`
// Number    string    `gorm:"not null;unique" json:"number"`
// Status    string    `json:"status"`
// Accrual   float64   `json:"accrual"`
// CreatedAt time.Time `json:"uploaded_at"`
// UpdatedAt time.Time `json:"-"`
// DeletedAt time.Time `json:"-"`
// UserID    uint      `json:"-"`
