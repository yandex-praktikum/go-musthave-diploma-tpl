package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	ID         uuid.UUID `json:"-"`
	Number     string    `json:"number"`
	Accrual    float32   `json:"accrual,omitempty"`
	Status     string    `json:"status"`
	UserID     uuid.UUID `json:"-"`
	UploadedAt time.Time `json:"uploaded_at"`
}
