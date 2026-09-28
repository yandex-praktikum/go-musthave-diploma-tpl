// Package money keeps loyalty points as cents to avoid float math.
package money

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Amount is a decimal money value stored as cents.
type Amount int64

// ErrInvalidAmount is returned when a decimal value cannot be parsed.
var ErrInvalidAmount = errors.New("invalid amount")

// Parse converts a decimal string into Amount.
func Parse(value string) (Amount, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ErrInvalidAmount
	}
	if strings.HasPrefix(value, "-") {
		return 0, ErrInvalidAmount
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrInvalidAmount
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}

	var cents int64
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) > 2 {
			return 0, ErrInvalidAmount
		}
		if len(fraction) == 1 {
			fraction += "0"
		}
		if len(fraction) == 0 {
			fraction = "00"
		}
		cents, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, ErrInvalidAmount
		}
	}

	return Amount(whole*100 + cents), nil
}

// String formats Amount as a JSON-friendly decimal number.
func (a Amount) String() string {
	sign := ""
	value := int64(a)
	if value < 0 {
		sign = "-"
		value = -value
	}

	whole := value / 100
	cents := value % 100
	switch {
	case cents == 0:
		return fmt.Sprintf("%s%d", sign, whole)
	case cents%10 == 0:
		return fmt.Sprintf("%s%d.%d", sign, whole, cents/10)
	default:
		return fmt.Sprintf("%s%d.%02d", sign, whole, cents)
	}
}

// MarshalJSON writes Amount as a JSON number.
func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalJSON reads Amount from a JSON number or string.
func (a *Amount) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*a = 0
		return nil
	}

	var raw string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &raw); err != nil {
			return ErrInvalidAmount
		}
	} else {
		raw = string(data)
	}

	value, err := Parse(raw)
	if err != nil {
		return err
	}
	*a = value
	return nil
}
