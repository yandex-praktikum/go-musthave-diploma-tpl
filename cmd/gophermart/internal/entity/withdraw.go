package entity

import (
	"time"
)

type Withdraw struct {
	UserID      string
	OrderNumber string
	Sum         float32
	ProcessedAt time.Time
}
