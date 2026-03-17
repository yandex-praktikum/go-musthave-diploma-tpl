package dto

import (
	"gmart/internal/domain"
	"time"
)

type OrderItem struct {
	Number     domain.OrderNumber `json:"number"`
	Status     domain.OrderStatus `json:"status"`
	Accrual    domain.Money       `json:"accrual"`
	UploadedAt time.Time          `json:"uploaded_at"`
}
