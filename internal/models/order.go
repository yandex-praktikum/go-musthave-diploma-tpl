package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ORDER_STATUS_NEW        = "NEW"
	ORDER_STATUS_PROCESSING = "PROCESSING"
	ORDER_STATUS_INVALID    = "INVALID"
	ORDER_STATUS_PROCESSED  = "PROCESSED"
)

type Order struct {
	ID         uuid.UUID `json:"-"`
	Number     string    `json:"number"`
	Accrual    uint      `json:"accrual,omitempty"`
	Status     string    `json:"status"`
	UserID     uuid.UUID `json:"-"`
	UploadedAt time.Time `json:"uploaded_at"`
}
