package dto

import (
	"gmart/internal/domain"
	"time"
)

type OrderItem struct {
	Number     domain.OrderNumber `json:"number"`
	Status     domain.OrderStatus `json:"status"`
	Accrual    domain.Amount       `json:"accrual,omitempty"`
	UploadedAt time.Time          `json:"uploaded_at"`
}
