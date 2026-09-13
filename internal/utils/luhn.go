// Package utils содержит вспомогательные утилиты: проверку номеров заказов
// по алгоритму Луна и работу с JWT-токенами.
package utils

// ValidLuhn проверяет номер заказа по алгоритму Луна.
//
// Возвращает true, если номер непустой, состоит только из цифр
// и проходит проверку по алгоритму Луна, иначе false.
func ValidLuhn(number string) bool {
	if number == "" {
		return false
	}

	var sum int
	var alternate bool

	// Идем справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
