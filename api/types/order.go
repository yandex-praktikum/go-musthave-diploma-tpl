package types

import (
	"github.com/google/uuid"
	"time"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
)

type OrderCreateResponse struct {
	Order *entities.Order `json:"order"`
}

type OrderResponse struct {
	UserID    uuid.UUID `json:"-"`
	Number    uint64    `json:"number,string"`
	Status    string    `json:"status"`
	Accrual   int64     `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
}

func NewOrderFromDomain(orderEntity *entities.Order) *OrderResponse {
	if orderEntity == nil {
		return nil
	}

	return &OrderResponse{
		UserID:    orderEntity.UserID,
		Number:    orderEntity.ID,
		Status:    orderEntity.Status,
		Accrual:   orderEntity.Accrual,
		CreatedAt: orderEntity.CreatedAt,
	}
}
