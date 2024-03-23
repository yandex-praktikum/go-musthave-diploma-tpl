package types

import (
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/A-Kuklin/gophermart/internal/domain/entities"
)

type BalanceResponse struct {
	Current     float64 `json:"current"`
	Withdrawals float64 `json:"withdrawn"`
}

type WithdrawCreateRequest struct {
	OrderID string  `json:"order"`
	Sum     float64 `json:"sum"`
}

func (c *WithdrawCreateRequest) ToDomain() (*entities.Withdraw, error) {
	orderID, err := strconv.ParseUint(c.OrderID, 0, 64)
	if err != nil {
		return nil, err
	}

	sum := int64(c.Sum * 100)

	withdraw := entities.Withdraw{
		OrderID: orderID,
		Amount:  sum,
	}

	return &withdraw, nil
}

type WithdrawResponse struct {
	UserID    uuid.UUID `json:"-"`
	OrderID   uint64    `json:"order,string"`
	Withdraw  int64     `json:"sum,omitempty"`
	CreatedAt time.Time `json:"processed_at"`
}

func NewWithdrawFromDomain(withdrawEntity *entities.Withdraw) *WithdrawResponse {
	if withdrawEntity == nil {
		return nil
	}

	return &WithdrawResponse{
		UserID:    withdrawEntity.UserID,
		OrderID:   withdrawEntity.OrderID,
		Withdraw:  withdrawEntity.Amount,
		CreatedAt: withdrawEntity.CreatedAt,
	}
}
