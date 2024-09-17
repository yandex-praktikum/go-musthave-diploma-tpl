package models

import "time"

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
	Number     string    `json:"order_id" db:"order_id"`
	Status     *string   `json:"status" db:"status"`
	Accrual    *float64  `json:"bonus,omitempty" db:"bonus"`
	UploadedAt time.Time `json:"created_in" db:"created_in"`
}

type ResponseAccrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Balance struct {
	Current  *float64 `json:"current"`
	Withdraw *int     `json:"withdraw"`
}

type Withdrawals struct {
	Order       string    `json:"order"`
	Sum         *float64  `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type statusKey string

const (
	NewOrder        statusKey = "NEW"        // Заказ загружен в систему, но не попал в обработку
	ProcessingOrder statusKey = "PROCESSING" // Вознаграждение за заказ рассчитывается
	InvalidOrder    statusKey = "INVALID"    // Система расчёта вознаграждений отказала в расчёте
	ProcessedOrder  statusKey = "PROCESSED"  // Данные по заказу проверены и информация о расчёте успешно получена
)
