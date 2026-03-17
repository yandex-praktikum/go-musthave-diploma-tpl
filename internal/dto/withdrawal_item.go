package dto

import (
	"gmart/internal/domain"
	"time"
)

// WithdrawalItem описывает одну запись в истории списаний
type WithdrawalItem struct {
	Order       domain.OrderNumber `json:"order"`
	Sum         domain.Money       `json:"sum"`
	ProcessedAt time.Time          `json:"processed_at"`
}
