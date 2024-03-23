package types

type Order struct {
	OrderID uint64
	Status  string
}

type AccrualResponse struct {
	OrderID uint64  `json:"order,string"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}
