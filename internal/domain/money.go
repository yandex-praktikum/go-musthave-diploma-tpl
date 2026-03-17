package domain

import (
	"bytes"
	"fmt"
	"strconv"
)

type Money uint64

// MarshalJSON конвертирует копейки в рубли (число с плавающей точкой) для JSON
func (m Money) MarshalJSON() ([]byte, error) {
	// Максимальный uint64 в рублях с точкой и 2 знаками копеек занимает 21 байт.
	buf := make([]byte, 0, 24)

	rub := uint64(m / 100)
	kop := uint64(m % 100)

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
func (m *Money) UnmarshalJSON(data []byte) error {
	// Убираем кавычки и пробелы
	d := bytes.Trim(data, "\" ")

	if len(d) == 0 || bytes.Equal(d, []byte("null")) {
		*m = 0
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
				return fmt.Errorf("money: multiple dots in %q", d)
			}
			foundDot = true
			continue
		}

		if c < '0' || c > '9' {
			return fmt.Errorf("money: invalid character %c", c)
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
					return fmt.Errorf("money: too much precision in %q", d)
				}
			}
		}
	}

	*m = Money(rub*100 + kop)
	return nil
}
