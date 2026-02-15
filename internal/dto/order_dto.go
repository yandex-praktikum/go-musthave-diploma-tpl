package dto

import "time"

type OrderInfo struct {
	Number     string    `json: "number"`
	Status     string    `json: "status"`
	Accrual    int       `json: "accrual"`
	UploadedAt time.Time `json: "uploaded_at"`
}

func (i *OrderInfo) IsEqual(other *AccrualCalculatorDTO) bool {
	return i.Status == other.Status
}

func (i *OrderInfo) Update(other *AccrualCalculatorDTO) {
	i.Status = other.Status
	i.Accrual = other.Accrual
}

type BalanceInfo struct {
	Current  float64 `json: "current"`
	Withdraw int     `json: "withdrawn"`
}
