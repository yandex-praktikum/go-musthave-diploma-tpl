package alg

import (
	"strings"
)

// IsValidLuhn проверяет строку с номером по алгоритму Луна
func IsValidLuhn(number string) bool {
	// Очистка от пробелов и дефисов (опционально)
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")

	// Проверка, что остались только цифры
	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}

	// Идём с конца строки к началу
	var sum int
	nDigits := len(number)
	parity := nDigits % 2 // Чётность длины определяет, какие цифры удваивать

	for i, r := range number {
		digit := int(r - '0')

		if i%2 == parity {
			digit *= 2
			// Если результат удвоения > 9, вычитаем 9 (или суммируем цифры результата: 12 -> 1+2=3, что то же самое что 12-9=3)
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	// Номер валиден, если сумма кратна 10
	return sum%10 == 0
}
