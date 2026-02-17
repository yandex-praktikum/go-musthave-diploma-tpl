package domain

import "time"

// OrderStatus статусы обработки расчётов
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"        // заказ загружен в систему, но не попал в обработку
	OrderStatusProcessing OrderStatus = "PROCESSING" // вознаграждение за заказ рассчитывается
	OrderStatusInvalid    OrderStatus = "INVALID"    // система расчёта вознаграждений отказала в расчёте
	OrderStatusProcessed  OrderStatus = "PROCESSED"  // данные по заказу проверены и информация о расчёте успешно получена
)

// User структура пользователя
type User struct {
	ID       string
	Login    string
	PassHash string
}

// Order структура заказа
type Order struct {
	ID         int64
	Number     string      // номер заказа
	UserID     string      // ID пользователя
	Status     OrderStatus // статус обработки
	Accrual    float64     // начисленные баллы
	UploadedAt time.Time   // время загрузки
}

// Balance структура баланса пользователя
type Balance struct {
	UserID    string
	Current   float64 // текущий баланс
	Withdrawn float64 // всего списано за всё время
}

// Withdrawal структура списания
type Withdrawal struct {
	ID          int64
	UserID      string
	OrderNumber string    // номер заказа, в счёт которого списываются баллы
	Sum         float64   // сумма списания
	ProcessedAt time.Time // время списания
}
