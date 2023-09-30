package models

import "time"

type Users struct {
	ID       int32   `gorm:"serial primary_key"`
	Login    string  `gorm:"not null"`
	Password string  `gorm:"not null"`
	Balance  float32 `gorm:"default 0"`
}

type Orders struct {
	ID            int32     `gorm:"serial primary_key"`
	Number        string    `gorm:"text not null"`
	UserLogin     string    `gorm:"text not null"`
	UploadedAt    time.Time `gorm:"timestamp with time zone default now() not null"`
	OperationType string    `gorm:"text default 'accrual'"`
	Status        string    `gorm:"not null default 'new'"`
	Amount        float32   `gorm:"default 0 not null"`
}

type User struct {
	Login    string `gorm:"column:login" json:"login"`
	Password string `gorm:"column:password" json:"password"`
	Balance  int    `gorm:"column:balance" json:"balance"`
}
