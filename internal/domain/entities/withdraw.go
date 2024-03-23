package entities

import (
	"time"

	"github.com/google/uuid"
)

type Withdraw struct {
	OrderID   uint64
	UserID    uuid.UUID
	Amount    int64
	CreatedAt time.Time
}
