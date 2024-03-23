package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusUnknown    = "UNKNOWN"
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	ID        uint64
	UserID    uuid.UUID
	Status    string
	Accrual   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
