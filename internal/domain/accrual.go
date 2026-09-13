package domain

// AccrualResponse представляет ответ от внешней системы расчёта баллов.
type AccrualResponse struct {
	// Order — номер заказа.
	Order string `json:"order"`
	// Status — статус расчёта: REGISTERED, INVALID, PROCESSING, PROCESSED.
	Status string `json:"status"`
	// Accrual — начисленные баллы (если рассчитаны).
	Accrual *float64 `json:"accrual,omitempty"`
}

// Статусы ответа от accrual-сервиса.
const (
	// AccrualStatusRegistered — заказ зарегистрирован, но не рассчитан.
	AccrualStatusRegistered = "REGISTERED"
	// AccrualStatusInvalid — заказ не принят к расчёту.
	AccrualStatusInvalid = "INVALID"
	// AccrualStatusProcessing — расчёт вознаграждения в процессе.
	AccrualStatusProcessing = "PROCESSING"
	// AccrualStatusProcessed — расчёт завершён.
	AccrualStatusProcessed = "PROCESSED"
)
