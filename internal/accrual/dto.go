package accrual

// OrderResponse — ответ системы начислений при 200 OK (GET /api/orders/{number}).
type OrderResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual *int   `json:"accrual,omitempty"`
}
