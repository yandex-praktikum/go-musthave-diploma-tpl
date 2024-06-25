package models

import "github.com/ShukinDmitriy/gophermart/internal/entities"

type AccrualOrderResponse struct {
	Order   string               `json:"order"`
	Status  entities.OrderStatus `json:"status"`
	Accrual float32              `json:"accrual"`
}
