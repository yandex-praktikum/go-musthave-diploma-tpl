package model

import (
	"encoding/json"
	"time"
)

const (
	CONFLICT    = "данный URL"
	ERRCONFLICT = "ERRCONFLICT"
	//  статус расчёта начисления из blackBox
	REGISTERED  = "REGISTERED"
	INVALID     = "INVALID"
	PROCESSING  = "PROCESSING"
	PROCESSED   = "PROCESSED"
	NEW         = "NEW"
	ERROR       = "ERROR"
	CALCULATION = "CALCULATION" // ДЛЯ НОВЫХ ЗАКАЗОВ
)

type User struct {
	Login     string
	PassHash  string
	OrderList map[int]*Order //[]*Order
}
type Order struct {
	OrderID int
	Status  string
	Created time.Time
	Accrual string
}
type AccrualRes struct {
	Order   string      `json:"order"`
	Status  string      `json:"status"`
	Accrual json.Number `json:"accrual"`
}
