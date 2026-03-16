package utils

import (
	"strconv"
	"strings"
)

// IsValidLuhn проверяет валидность строки с помощью алгоритма Луна
// Принимает строку, содержащую цифры (допускаются пробелы)
// Возвращает true если номер валиден, иначе false
func IsValidLuhn(input string) bool {
	cleaned := strings.ReplaceAll(input, " ", "")

	if len(cleaned) < 2 {
		return false
	}

	for _, r := range cleaned {
		if r < '0' || r > '9' {
			return false
		}
	}

	sum := 0
	isSecond := false

	for i := len(cleaned) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(cleaned[i]))

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
