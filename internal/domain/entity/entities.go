package entity

// User модель пользователя
type User struct {
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Login     string `json:"login"`
	Password  string `json:"-"`
	IsActive  bool   `json:"is_active"`
}

// OrderStatus тип для статуса заказа
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

// Order модель заказа
type Order struct {
	ID        int         `json:"id"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at,omitempty"`
	Number    string      `json:"number"`
	Status    OrderStatus `json:"status"`
	Accrual   *float64    `json:"accrual,omitempty"`
	UserID    int         `json:"user_id"`
}

// OrderWithUser заказ с информацией о пользователе
type OrderWithUser struct {
	Order
	User User `json:"user"`
}

// UserBalance информация о балансе пользователя
type UserBalance struct {
	//UserID    int     `json:"user_id"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdraw списание средств
type Withdraw struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
