package accrual

// accrualOrder is the order status response payload from the accrual system.
type accrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
