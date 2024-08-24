package models

import (
	"time"

	"github.com/google/uuid"
)

type Withdraw struct {
	ID          uuid.UUID `json:"id"`
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
	UserID      uuid.UUID `json:"-"`
}
