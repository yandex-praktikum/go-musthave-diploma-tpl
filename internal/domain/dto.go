package domain

import "time"

// RegisterRequest — запрос на регистрацию пользователя.
type RegisterRequest struct {
	// Login — логин пользователя.
	Login string `json:"login"`
	// Password — пароль пользователя в открытом виде.
	Password string `json:"password"`
}

// LoginRequest — запрос на аутентификацию пользователя.
type LoginRequest struct {
	// Login — логин пользователя.
	Login string `json:"login"`
	// Password — пароль пользователя в открытом виде.
	Password string `json:"password"`
}

// WithdrawRequest — запрос на списание баллов.
type WithdrawRequest struct {
	// Order — номер заказа, за который производится списание.
	Order string `json:"order"`
	// Sum — сумма списания.
	Sum float64 `json:"sum"`
}

// OrderResponse — элемент ответа со списком заказов пользователя.
type OrderResponse struct {
	// Number — номер заказа.
	Number string `json:"number"`
	// Status — статус обработки заказа.
	Status string `json:"status"`
	// Accrual — начисленные баллы (отсутствует, если не рассчитаны).
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt — время загрузки заказа.
	UploadedAt time.Time `json:"uploaded_at"`
}

// BalanceResponse — ответ с балансом пользователя.
type BalanceResponse struct {
	// Current — текущий баланс.
	Current float64 `json:"current"`
	// Withdrawn — сумма всех списаний за всё время.
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawalResponse — элемент ответа с историей списаний.
type WithdrawalResponse struct {
	// Order — номер заказа, за который произведено списание.
	Order string `json:"order"`
	// Sum — сумма списания.
	Sum float64 `json:"sum"`
	// ProcessedAt — время списания.
	ProcessedAt time.Time `json:"processed_at"`
}
