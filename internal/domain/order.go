package domain

import "time"

// Order представляет заказ, загруженный пользователем.
type Order struct {
	// Number — уникальный номер заказа.
	Number string `json:"number"`
	// UserID — идентификатор пользователя, загрузившего заказ.
	UserID int64 `json:"user_id"`
	// Status — статус обработки заказа: NEW, PROCESSING, INVALID, PROCESSED.
	Status string `json:"status"`
	// Accrual — начисленные баллы (nil, если не рассчитаны).
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt — время загрузки заказа.
	UploadedAt time.Time `json:"uploaded_at"`
	// UpdatedAt — время последнего обновления заказа.
	UpdatedAt time.Time `json:"updated_at"`
}

// Статусы заказа.
const (
	// OrderStatusNew — заказ загружен, но не передан в расчёт.
	OrderStatusNew = "NEW"
	// OrderStatusProcessing — вознаграждение рассчитывается.
	OrderStatusProcessing = "PROCESSING"
	// OrderStatusInvalid — система отказала в расчёте.
	OrderStatusInvalid = "INVALID"
	// OrderStatusProcessed — расчёт завершён.
	OrderStatusProcessed = "PROCESSED"
)
