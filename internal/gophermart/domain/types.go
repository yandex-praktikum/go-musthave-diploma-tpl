package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type OrderStatus string

type OrderNumber string

func (os *OrderNumber) FromString(input string) {
	// TODO - проверка по алгоритму луна
	*os = OrderNumber(input)
}

const (
	OrderStratusNew        OrderStatus = "NEW"
	OrderStratusProcessing OrderStatus = "PROCESSING"
	OrderStratusInvalid    OrderStatus = "INVALID"
	OrderStratusProcessed  OrderStatus = "PROCESSED"
)

type AccrualStatus string

const (
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

type RegisterData struct {
	Login    string
	Password string
}

type LoginData struct {
	Login    string
	Password string
}

type TokenString string

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type AuthData struct {
	UserID int
}

type OrderData struct {
	Number     OrderNumber `json:"number"`
	Status     OrderStatus `json:"status,omitempty"`
	Accrual    int         `json:"accrual,omitempty"`
	UploadedAt int         `json:"uploaded_at,omitempty"`
}

type Balance struct {
	Current   float64 `json:"current"`
	WithDrawn float64 `json:"withdrawn"`
}

type WithdrawData struct {
	Order       OrderNumber `json:"order"`
	Sum         int         `json:"sum"`
	ProcessedAt time.Time   `json:"processed_at,omitempty"`
}

type AccrualData struct {
	Number  OrderNumber   `json:"number"`
	Status  AccrualStatus `json:"status"`
	Accrual *int          `jsin:"accrual,omitempty"`
}
