package models

import "encoding/json"

type (
	AccrualResponse struct {
		Order   string      `json:"order"`
		Status  string      `json:"status"`
		Accrual json.Number `json:"accrual"` // json.Number хранит число как строку "172.32", будем потом переводить в целые копейки
	}
)
