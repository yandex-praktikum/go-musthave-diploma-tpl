package models

import "time"

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithDrawRequest struct {
	NumberOrder string  `json:"order"`
	Sum         float64 `json:"sum"`
}

type UserWithDraw struct {
	NumberOrder string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
