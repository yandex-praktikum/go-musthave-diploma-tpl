package handler

// RegisterRequest — тело запроса регистрации.
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginRequest — тело запроса логина.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// OrderItem — элемент списка заказов.
type OrderItem struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

// BalanceResponse — ответ GET /api/user/balance.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawRequest — тело POST /api/user/balance/withdraw.
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// WithdrawalItem — элемент списка списаний GET /api/user/withdrawals.
type WithdrawalItem struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
