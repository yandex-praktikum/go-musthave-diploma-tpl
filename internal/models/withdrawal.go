package models

import "time"

// Withdrawal — запись о списании баллов (таблица withdrawals).
type Withdrawal struct {
	ID          int64
	UserID      int64
	Order       string
	Sum         int64
	ProcessedAt time.Time
}
