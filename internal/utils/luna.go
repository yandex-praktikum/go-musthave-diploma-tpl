package utils

import "strconv"

// Функция, реализующая алгоритм Луна для проверки корректности номера заказа
func IsLunaValid(number string) bool {
	var sum int
	// Перебираем цифры с конца к началу
	for i, r := range number {
		digit, err := strconv.Atoi(string(r))
		if err != nil {
			// Если символ не является цифрой, возвращаем false
			return false
		}

		// Если индекс четный, цифру нужно удвоить
		if (len(number)-1-i)%2 == 1 {
			digit *= 2
			if digit > 9 {
				// Если результат больше 9, складываем цифры
				digit -= 9
			}
		}

		sum += digit
	}

	// Проверяем, кратна ли сумма 10
	return sum%10 == 0
}
