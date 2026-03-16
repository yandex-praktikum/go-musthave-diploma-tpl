package entity

import "time"

type Withdrawal struct {
	ID          int
	UserID      int
	OrderID     string
	Sum         float64
	ProcessedAt time.Time
}
