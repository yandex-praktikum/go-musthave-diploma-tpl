package models

import "time"

type Users struct {
	ID        int32   `gorm:"serial primary_key"`
	Login     string  `gorm:"not null"`
	Password  string  `gorm:"not null"`
	Balance   float32 `gorm:"default 0"`
	Withdrawn float32 `gorm:"default 0"`
}

type Orders struct {
	ID            int32     `gorm:"primaryKey" json:"-"`
	Number        string    `gorm:"not null;unique" json:"number"`
	UserLogin     string    `json:"-"`
	UploadedAt    time.Time `json:"uploaded_at"`
	Accrual       float32   `json:"accrual"`
	Status        string    `json:"status"`
	OperationType string    `json:"-"`
}

type Balance struct {
	Login     string  `json:"user_login,omitempty"`
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdraw struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

type OrdersWithdrawn struct {
	Number       string    `json:"order"`
	Sum          float32   `json:"sum"`
	Processed_At time.Time `json:"processed_at"`
}
