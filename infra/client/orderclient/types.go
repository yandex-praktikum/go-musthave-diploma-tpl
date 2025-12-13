package orderclient

// OrderStatus представляет статус заказа
type OrderStatus string

const (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusProcessed  OrderStatus = "PROCESSED"
)

// OrderResponse представляет ответ API о заказе
type OrderResponse struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual *float64    `json:"accrual,omitempty"`
}

// IsValid проверяет валидность статуса
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusRegistered, StatusInvalid, StatusProcessing, StatusProcessed:
		return true
	default:
		return false
	}
}

// HasAccrual проверяет, есть ли начисление баллов
func (r *OrderResponse) HasAccrual() bool {
	return r.Accrual != nil
}

// Config конфигурация клиента
type Config struct {
	// BaseURL базовый URL API
	BaseURL string
	// Timeout таймаут запроса в секундах
	Timeout int
	// MaxRetries максимальное количество повторных попыток
	MaxRetries int
	// RetryDelay задержка между попытками в секундах
	RetryDelay int
	// UserAgent заголовок User-Agent
	UserAgent string
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		BaseURL:    "http://localhost:8080",
		Timeout:    10,
		MaxRetries: 3,
		RetryDelay: 1,
		UserAgent:  "OrderClient/1.0",
	}
}
