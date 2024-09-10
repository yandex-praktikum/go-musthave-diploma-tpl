package models

import "time"

type Loyalty struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OrderID     string    `json:"order_id"`
	Bonus       int       `json:"bonus"`
	OrderStatus string    `json:"order_status"`
	IsDeleted   bool      `json:"is_deleted"`
	CreatedIn   time.Time `json:"created_in"`
}

type OrdersUser struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}
