package entity

import "time"

type JobStoreRow struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	NextTimeExecute time.Time `json:"next_time_execute"`
	Count           int       `json:"count"`
	Executed        bool      `json:"executed"`
	Parameters      string    `json:"parameters"`
}

type CheckOrderStatusParameters struct {
	OrderNumber string `json:"order_number"`
}
