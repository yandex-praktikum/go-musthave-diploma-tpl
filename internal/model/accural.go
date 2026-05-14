// Модель при обращении в сервис Accrual
package model

type AccuralResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

type AccuralRequest struct {
	Order string `json:"order"`
}
