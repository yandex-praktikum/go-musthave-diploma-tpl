package models

import (
	"fmt"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

type GetOrdersResponse struct {
	Number     string               `json:"number"`
	Status     entities.OrderStatus `json:"status"`
	Accrual    float32              `json:"accrual,omitempty"`
	UploadedAt JSONTime             `json:"uploaded_at"`
}
