package domain

// Balance представляет баланс баллов лояльности пользователя.
type Balance struct {
	// UserID — идентификатор пользователя.
	UserID int64 `json:"user_id"`
	// Current — текущий баланс.
	Current float64 `json:"current"`
	// Withdrawn — сумма всех списаний за всё время.
	Withdrawn float64 `json:"withdrawn"`
}
