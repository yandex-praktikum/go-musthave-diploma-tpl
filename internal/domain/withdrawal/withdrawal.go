package withdrawal

import "time"

type Withdrawal struct {
	Id          string    `json:"-"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	PricessedAt time.Time `json:"processed_at"`
	UserId      string    `json:"-"`
}
