package domain

import "time"

// Withdrawal представляет запись о списании баллов.
type Withdrawal struct {
	// ID — идентификатор записи о списании.
	ID int64 `json:"id"`
	// UserID — идентификатор пользователя, выполнившего списание.
	UserID int64 `json:"user_id"`
	// OrderNumber — номер заказа, за который произведено списание.
	OrderNumber string `json:"order"`
	// Sum — сумма списания.
	Sum float64 `json:"sum"`
	// ProcessedAt — время списания.
	ProcessedAt time.Time `json:"processed_at"`
}
