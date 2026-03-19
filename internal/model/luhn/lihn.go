package luhn

import "gmart/internal/domain"

// IsValid проверяет строку на соответствие алгоритму Луна.
// Работает с произвольной длиной номера заказа.
func IsValid(number domain.OrderNumber) bool {
	var sum int
	parity := len(number) % 2

	for i, r := range number {
		// Превращаем руну в число
		digit := int(r - '0')
		if digit < 0 || digit > 9 {
			return false // В строке не число
		}

		// Удваиваем каждое второе число (начиная с левого края)
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	// Номер валиден, если сумма кратна 10
	return sum%10 == 0
}
