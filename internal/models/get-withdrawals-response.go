package models

type GetWithdrawalsResponse struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt *JSONTime `json:"processed_at"`
}
