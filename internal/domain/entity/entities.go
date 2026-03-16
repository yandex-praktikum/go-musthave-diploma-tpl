package entity

// ==================== Пользователи ====================

// User модель пользователя
type User struct {
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Login     string `json:"login"`
	Password  string `json:"-"` // Поле исключено из JSON для безопасности
	IsActive  bool   `json:"is_active"`
}

// UserWithOrders представляет пользователя с его заказами
type UserWithOrders struct {
	User   User    `json:"user"`
	Orders []Order `json:"orders"`
}

// ==================== Заказы ====================

// OrderStatus тип для статуса заказа
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"        // Новый заказ
	OrderStatusProcessing OrderStatus = "PROCESSING" // В обработке
	OrderStatusInvalid    OrderStatus = "INVALID"    // Невалидный
	OrderStatusProcessed  OrderStatus = "PROCESSED"  // Обработан
)

// Order модель заказа
type Order struct {
	ID          int         `json:"id"`
	UploadedAt  string      `json:"uploaded_at"`
	ProcessedAt string      `json:"processed_at,omitempty"`
	Number      string      `json:"number"`
	Status      OrderStatus `json:"status"`
	Accrual     *float64    `json:"accrual,omitempty"`
	UserID      int         `json:"user_id"`
}

// OrderWithUser представляет заказ с информацией о пользователе
type OrderWithUser struct {
	Order
	User User `json:"user"`
}

// ==================== Баланс и финансы ====================

// UserBalance информация о балансе пользователя
type UserBalance struct {
	UserID    int     `json:"user_id"`
	Current   float64 `json:"current"`   // Текущий баланс
	Withdrawn float64 `json:"withdrawn"` // Сумма списаний
}

// Withdrawal списание средств
type Withdrawal struct {
	UserID      int     `json:"user_id"`
	Order       string  `json:"order"`        // Номер заказа списания
	Sum         float64 `json:"sum"`          // Сумма списания
	ProcessedAt string  `json:"processed_at"` // Время обработки
}

// ==================== Статистика и сводки ====================

// BalanceSummary сводка по балансу
type BalanceSummary struct {
	TotalEarned    float64 `json:"total_earned"`    // Всего заработано
	TotalSpent     float64 `json:"total_spent"`     // Всего потрачено
	CurrentBalance float64 `json:"current_balance"` // Текущий баланс
}

// UserStats статистика пользователя
type UserStats struct {
	UserID           int     `json:"user_id"`
	TotalOrders      int     `json:"total_orders"`      // Всего заказов
	ProcessedOrders  int     `json:"processed_orders"`  // Обработанных заказов
	NewOrders        int     `json:"new_orders"`        // Новых заказов
	ProcessingOrders int     `json:"processing_orders"` // В обработке
	InvalidOrders    int     `json:"invalid_orders"`    // Невалидных заказов
	TotalAccrual     float64 `json:"total_accrual"`     // Общая сумма начислений
	TotalWithdrawn   float64 `json:"total_withdrawn"`   // Общая сумма списаний
}

// OrdersSummary сводка по заказам
type OrdersSummary struct {
	Total           int     `json:"total"`            // Всего заказов
	NewCount        int     `json:"new_count"`        // Новых
	ProcessingCount int     `json:"processing_count"` // В обработке
	ProcessedCount  int     `json:"processed_count"`  // Обработанных
	InvalidCount    int     `json:"invalid_count"`    // Невалидных
	TotalAccrual    float64 `json:"total_accrual"`    // Общая сумма начислений
}

// WithdrawalsSummary сводка по списаниям
type WithdrawalsSummary struct {
	TotalWithdrawals int     `json:"total_withdrawals"` // Всего списаний
	TotalAmount      float64 `json:"total_amount"`      // Общая сумма
	UniqueUsers      int     `json:"unique_users"`      // Уникальных пользователей
}

// UserWithdrawalsSummary сводка по списаниям пользователя
type UserWithdrawalsSummary struct {
	UserID          int     `json:"user_id"`
	WithdrawalCount int     `json:"withdrawal_count"`           // Количество списаний
	TotalAmount     float64 `json:"total_amount"`               // Общая сумма списаний
	FirstWithdrawal string  `json:"first_withdrawal,omitempty"` // Первое списание
	LastWithdrawal  string  `json:"last_withdrawal,omitempty"`  // Последнее списание
}
