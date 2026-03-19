package domain

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
)

type Amount uint64

// Schema сообщает Huma, как этот тип должен выглядеть в OpenAPI
func (a Amount) Schema(r huma.Registry) *huma.Schema {
	multipleOf := 0.01
	return &huma.Schema{
		Type:        "number",
		Format:      "double",
		MultipleOf:  &multipleOf,
		Description: "Денежная сумма в формате 1.23",
		Examples:    []any{123.45},
	}
}

// MarshalJSON конвертирует копейки в рубли (число с плавающей точкой) для JSON
func (a Amount) MarshalJSON() ([]byte, error) {
	// Максимальный uint64 в рублях с точкой и 2 знаками копеек занимает 21 байт.
	buf := make([]byte, 0, 24)

	rub := uint64(a / 100)
	kop := uint64(a % 100)

	// Добавляем рубли
	buf = strconv.AppendUint(buf, rub, 10)

	// Добавляем точку
	buf = append(buf, '.')

	// Добавляем копейки с ведущим нулем (всегда 2 знака)
	if kop < 10 {
		buf = append(buf, '0')
	}
	buf = strconv.AppendUint(buf, kop, 10)

	return buf, nil
}

// UnmarshalJSON принимает рубли из JSON и сохраняет как копейки
func (a *Amount) UnmarshalJSON(data []byte) error {
	// Убираем кавычки и пробелы
	d := bytes.Trim(data, "\" ")

	if len(d) == 0 || bytes.Equal(d, []byte("null")) {
		*a = 0
		return nil
	}

	var (
		rub      uint64
		kop      uint64
		foundDot bool
		fracLen  int
	)

	for _, c := range d {
		if c == '.' {
			if foundDot {
				return fmt.Errorf("amount: multiple dots in %q", d)
			}
			foundDot = true
			continue
		}

		if c < '0' || c > '9' {
			return fmt.Errorf("amount: invalid character %c", c)
		}

		digit := uint64(c - '0')

		if !foundDot {
			rub = rub*10 + digit
		} else {
			fracLen++
			switch fracLen {
			case 1:
				kop = digit * 10
			case 2:
				kop += digit
			default:
				// Толерантность только к нулям. Лишние не нули — ошибка.
				if digit != 0 {
					return fmt.Errorf("amount: too much precision in %q", d)
				}
			}
		}
	}

	*a = Amount(rub*100 + kop)
	return nil
}
