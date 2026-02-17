package dto

import "time"

type WithdrawRequest struct {
	Order string `json: "order"`
	Sum   int    `json: "sum"`
}

type WithdrawInfo struct {
	Order       string    `json: "order"`
	Sum         int       `json: "sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func NewWithdrawInfo(req WithdrawRequest) *WithdrawInfo {
	return &WithdrawInfo{
		Order:       req.Order,
		Sum:         req.Sum,
		ProcessedAt: time.Now(),
	}
}

type AccrualCalculatorDTO struct {
	Order   string `json: "order"`
	Status  string `json: "status"`
	Accrual int    `json: "accrual"`
}

func (dto *AccrualCalculatorDTO) IsEqual(other AccrualCalculatorDTO) bool {
	return dto.Order == other.Order && dto.Status == other.Status && dto.Accrual == other.Accrual
}
