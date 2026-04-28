package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const (
	OrderNewStatus       = "NEW"
	OrderInProcessStatus = "PROCESSING"
	OrderInvalidStatus   = "INVALID"
	OrderProcessedStatus = "PROCESSED"
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderExists             = errors.New("order exists")
	ErrOrderBelongsToOtherUser = errors.New("order belongs to other user")

	validStatuses = map[string]bool{
		OrderNewStatus:       true,
		OrderInProcessStatus: true,
		OrderInvalidStatus:   true,
		OrderProcessedStatus: true,
	}
)

type (
	/*
		Объект пользователя системы
	*/
	Order struct {
		ID         int64     `json:"id"`
		UserID     string    `json:"user_id"`
		Status     string    `json:"status"`
		Accrual    int       `json:"accrual"`
		UploadedAt time.Time `json:"uploaded,omitempty"`
		UpdatedAt  time.Time `json:"updated,omitempty"`
	}

	/*
		Интерфейс объекта пользователя системы
	*/
	OrderInterface interface {
		ToJSON() ([]byte, error)
		Validate() error
	}
)

/*
Конструктор объекта заказа.
*/
func NewOrder(idStr, userID, status string) (*Order, error) {
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return &Order{ID: idInt, UserID: userID, Status: status}, nil
}

/*
Валидация заказа
*/
func (m *Order) Validate() error {
	switch {
	case m.UserID == "":
		return errors.New("UserID can't be empty")
	case m.ID == 0:
		return errors.New("ID can't be equal to zero")
	case !validStatuses[m.Status]:
		return errors.New("wrong status value")
	case m.Accrual < 0:
		return errors.New("accrual must be a positive or zero")
	}
	return nil
}

/*
Сериализует объект пользователя в строку JSON
Может вернуть ошибку, коли та случится.
*/
func (m *Order) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
