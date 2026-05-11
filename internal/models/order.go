package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	OrderNewStatus        = "NEW"
	OrderRegisteredStatus = "REGISTERED"
	OrderInProcessStatus  = "PROCESSING"
	OrderInvalidStatus    = "INVALID"
	OrderProcessedStatus  = "PROCESSED"
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderExists             = errors.New("order exists")
	ErrOrderBelongsToOtherUser = errors.New("order belongs to other user")
	ErrOrderWrongFormat        = errors.New("wrong order format")
	ErrORderWrongStatus        = errors.New("wrong status value")

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
		ID         string    `json:"number"`
		UserID     string    `json:"-"`
		Status     string    `json:"status"`
		Accrual    Accrual   `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded"`
		UpdatedAt  time.Time `json:"-"`
	}

	/*
		Интерфейс объекта пользователя системы
	*/
	OrderInterface interface {
		ToJSON() ([]byte, error)
		Validate() error
	}

	OrdersList []*Order
)

/*
Конструктор объекта заказа.
*/
func NewOrder(id, userID, status string) (*Order, error) {
	return &Order{ID: id, UserID: userID, Status: status}, nil
}

/*
Валидация заказа
*/
func (m *Order) Validate() error {
	switch {
	case m.UserID == "":
		return errors.New("UserID can't be empty")
	case m.ID == "":
		return errors.New("ID can't be equal to zero")
	case !validStatuses[m.Status]:
		return ErrORderWrongStatus
	case m.Accrual < 0:
		return errors.New("accrual must be a positive or zero")
	}
	return nil
}

// Обновляем значение поля Status
func (m *Order) SetStatus(status string) error {
	if !validStatuses[status] {
		return ErrORderWrongStatus
	}
	m.Status = status
	return nil
}

// Обновляем значение поля Status
func (m *Order) SetAccrual(accrualStr string) error {
	accrualInt, err := m.converStringAccrualToInt(accrualStr)
	if err != nil {
		return err
	}

	m.Accrual = Accrual(accrualInt)

	return nil
}

/*
Сериализует объект пользователя в строку JSON
Может вернуть ошибку, коли та случится.
*/
func (m *Order) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Order) converStringAccrualToInt(accrualStr string) (int, error) {
	parts := strings.Split(accrualStr, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid number format: %s", accrualStr)
	}

	intPart, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("failed to parse integer part: %w", err)
	}

	result := intPart * 100

	if len(parts) == 2 {
		fracStr := parts[1]

		switch len(fracStr) {
		case 1:
			// "172.3" -> "30" (дописываем 0)
			fracStr += "0"
		case 2:
			// "172.32" -> "32" (оставляем как есть)
		default:
			// "172.321" -> "32" (обрезаем лишнее), что не очень хорошо
			fracStr = fracStr[:2]
		}

		fracPart, err := strconv.Atoi(fracStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse fractional part: %w", err)
		}

		if intPart < 0 {
			result -= fracPart
		} else {
			result += fracPart
		}
	}

	return result, nil
}

func (a *Accrual) Int() int {
	return int(*a)
}

// Кастомный маршаллер для переовд денег в копейках строки в строку с рублями и копейками
func (a Accrual) MarshalJSON() ([]byte, error) {
	if a == 0 {
		return []byte("0"), nil
	}

	val := int(a)
	sign := ""
	if val < 0 {
		sign = "-" // а вдруг будет отрицатлеьный баланс?
		val = -val
	}

	s := strconv.Itoa(val)

	// Дополняем нулями слева, если длина меньше 3 (например, 5 -> 005)
	if len(s) < 3 {
		s = strings.Repeat("0", 3-len(s)) + s
	}

	// Вставляем точку: отделяем последние 2 символа
	// "17348" -> "173" + "." + "48"
	// "005"   -> "0"   + "." + "05"
	intPart := s[:len(s)-2]
	fracPart := s[len(s)-2:]

	// Собираем результат.
	result := sign + intPart + "." + fracPart
	return []byte(result), nil
}
