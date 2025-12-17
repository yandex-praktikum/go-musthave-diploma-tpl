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
