package orders

import (
	"errors"
	"time"
)

type Order struct {
	Number     string    `json:"number" validate:"required"`
	Status     string    `json:"status" validate:"required"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserId     string    `json:"-"`
}

var (
	ErrInvalidNumber = errors.New("invalid number")
	//если заказ был уже добавлен ранее пользователем
	ErrExistForUser = errors.New("already exist")
	//если заказ был добавлен другим пользователем
	ErrForeignForUser = errors.New("already exist")
	//если пытаемся добавить заказ, который уже существует но проскочили проверки ErrExistForUser и ErrForeignForUser
	ErrAlreadyExist   = errors.New("already exist")
	ErrNotFound       = errors.New("not found")
	ErrUnexpectedFlow = errors.New("order create error")
)

type Status string

const (
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
)
