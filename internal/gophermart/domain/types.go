package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

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

type RegistrationData struct {
	Login    string
	Password string
}

type AuthentificationData struct {
	Login    string
	Password string
}

type LoginData struct {
	UserID int
	Login  string
	Hash   string
	Salt   string
}

type AuthData struct {
	UserID int
}

type OrderData struct {
	Number     OrderNumber `json:"number"`
	Status     OrderStatus `json:"status,omitempty"`
	Accrual    *float64    `json:"accrual,omitempty"`
	UploadedAt RFC3339Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	WithDrawn float64 `json:"withdrawn"`
}

type WithdrawData struct {
	Order OrderNumber `json:"order"`
	Sum   float64     `json:"sum"`
}

type WithdrawalsData struct {
	Order       OrderNumber `json:"order"`
	Sum         float64     `json:"sum"`
	ProcessedAt RFC3339Time `json:"processed_at,omitempty"`
}

type AccrualData struct {
	Number  OrderNumber   `json:"number"`
	Status  AccrualStatus `json:"status"`
	Accrual *float64      `jsin:"accrual,omitempty"`
}

type OrderStatus string

type OrderNumber string

type RFC3339Time time.Time

var _ json.Marshaler = (*RFC3339Time)(nil)
var _ json.Unmarshaler = (*RFC3339Time)(nil)

func (r RFC3339Time) MarshalJSON() ([]byte, error) {
	t := (time.Time)(r)
	return []byte(fmt.Sprintf(`"%s"`, t.Format(time.RFC3339))), nil
}

func (r *RFC3339Time) UnmarshalJSON(b []byte) error {
	val := strings.Trim(string(b), `"`)
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return err
	}
	*r = *TimePtr(t)
	return nil
}

func (n *OrderNumber) UnmarshalJSON(b []byte) error {
	val := strings.Trim(string(b), `"`)
	if !CheckLuhn(val) {
		return fmt.Errorf("%w: %v wrong value", ErrDataFormat, val)
	}
	*n = OrderNumber(val)
	return nil
}

func (s *AccrualStatus) UnmarshalJSON(b []byte) error {
	val := strings.Trim(string(b), `"`)

	// Проверяем что пришло от системы расчета баллов

	switch AccrualStatus(val) {
	case AccrualStatusRegistered:
		*s = AccrualStatusRegistered
		return nil

	case AccrualStatusInvalid:
		*s = AccrualStatusInvalid
		return nil

	case AccrualStatusProcessing:
		*s = AccrualStatusProcessing
		return nil

	case AccrualStatusProcessed:
		*s = AccrualStatusProcessed
		return nil
	default:
		return fmt.Errorf("%w: unexpected accrualStatus %s", ErrDataFormat, val)
	}
}

func CheckLuhn(on string) bool {
	sum := 0
	nDigits := len(on)
	parity := nDigits % 2
	for id, c := range on {
		digit := int(c - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if id%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

func TimePtr(v time.Time) *RFC3339Time {
	r := RFC3339Time(v)
	return &r
}

func Float64Ptr(v float64) *float64 {
	return &v
}
