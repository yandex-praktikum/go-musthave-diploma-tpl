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
	UserID    int     `json:"user_id"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal списание средств
type Withdrawal struct {
	UserID      int     `json:"user_id"`
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type BalanceSummary struct {
	TotalEarned    float64 `json:"total_earned"`
	TotalSpent     float64 `json:"total_spent"`
	CurrentBalance float64 `json:"current_balance"`
}

type UserStats struct {
	UserID           int     `json:"user_id"`
	TotalOrders      int     `json:"total_orders"`
	ProcessedOrders  int     `json:"processed_orders"`
	NewOrders        int     `json:"new_orders"`
	ProcessingOrders int     `json:"processing_orders"`
	InvalidOrders    int     `json:"invalid_orders"`
	TotalAccrual     float64 `json:"total_accrual"`
	TotalWithdrawn   float64 `json:"total_withdrawn"`
}

type UserWithOrders struct {
	User   User    `json:"user"`
	Orders []Order `json:"orders"`
}

// Структура для сводки по заказам (order.go)
type OrdersSummary struct {
	Total           int     `json:"total"`
	NewCount        int     `json:"new_count"`
	ProcessingCount int     `json:"processing_count"`
	ProcessedCount  int     `json:"processed_count"`
	InvalidCount    int     `json:"invalid_count"`
	TotalAccrual    float64 `json:"total_accrual"`
}

// Структура для сводки по списаниям (withdrawal.go)
type WithdrawalsSummary struct {
	TotalWithdrawals int     `json:"total_withdrawals"`
	TotalAmount      float64 `json:"total_amount"`
	UniqueUsers      int     `json:"unique_users"`
}

// Структура для сводки по списаниям пользователя (withdrawal.go)
type UserWithdrawalsSummary struct {
	UserID          int     `json:"user_id"`
	WithdrawalCount int     `json:"withdrawal_count"`
	TotalAmount     float64 `json:"total_amount"`
	FirstWithdrawal string  `json:"first_withdrawal,omitempty"`
	LastWithdrawal  string  `json:"last_withdrawal,omitempty"`
}
