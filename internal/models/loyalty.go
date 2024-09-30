package models

import "time"

const (
	NewOrder        = "NEW"        // Заказ загружен в систему, но не попал в обработку
	ProcessingOrder = "PROCESSING" // Вознаграждение за заказ рассчитывается
	InvalidOrder    = "INVALID"    // Система расчёта вознаграждений отказала в расчёте
	ProcessedOrder  = "PROCESSED"  // Данные по заказу проверены и информация о расчёте успешно получена
)

type Loyalty struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OrderID     string    `json:"order_id"`
	Bonus       float64   `json:"bonus"`
	OrderStatus string    `json:"order_status"`
	CreatedIn   time.Time `json:"created_in"`
	Withdraw    float64   `json:"withdraw"`
	ProcessedAt time.Time `json:"processed_at"`
}

type OrdersUser struct {
	Number     string    `json:"number" db:"order_id"`
	Status     *string   `json:"status" db:"status"`
	Accrual    *float64  `json:"accrual,omitempty" db:"bonus"`
	UploadedAt time.Time `json:"uploaded_at" db:"created_in"`
}

type ResponseAccrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

type Balance struct {
	Current  float32 `json:"current"`
	Withdraw float32 `json:"withdrawn"`
}

type Withdrawals struct {
	Order       string     `json:"order"`
	Sum         *float32   `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at"`
}
