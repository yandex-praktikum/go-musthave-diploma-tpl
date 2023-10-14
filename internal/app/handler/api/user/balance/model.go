package balance

import "time"

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn int     `json:"withdrawn"`
}

type Withdraw struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

type Withdrawals struct {
	Order       string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
