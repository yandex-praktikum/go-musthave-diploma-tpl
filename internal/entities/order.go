package entities

import (
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	gorm.Model
	Number  string      `json:"number" gorm:"type:varchar"`
	UserID  uint        `json:"user_id"`
	Status  OrderStatus `json:"status"`
	Accrual float32     `json:"accrual"`
}
