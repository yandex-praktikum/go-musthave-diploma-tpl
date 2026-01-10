package model

import "time"

const (
	CONFLICT    = "данный URL"
	ERRCONFLICT = "ERRCONFLICT"
	//  статус расчёта начисления из blackBox
	REGISTERED = "REGISTERED"
	INVALID    = "INVALID"
	PROCESSING = "PROCESSING"
	PROCESSED  = "PROCESSED"
	NEW        = "NEW"
)

type AliasFullCore struct {
	Alias       string `json:"alias"`
	OriginalURL string `json:"original_url"`
	CorrID      string `json:"correlation_id"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

type User struct {
	Login     string
	PassHash  string
	OrderList []*Order
}
type Order struct {
	OrderID string
	Status  string
	Created time.Time
	Accural string
}
