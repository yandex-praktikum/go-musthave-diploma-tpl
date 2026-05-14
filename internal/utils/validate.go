// Алгорит Moon для проверки номера счета

package utils

func MoonAlgorithm(order string) bool {
	if order == "" {
		return false
	}
	var sum int
	var isEven bool

	for i := len(order) - 1; i >= 0; i-- {
		digit := int(order[i] - '0')

		if digit < 0 || digit > 9 {
			return false
		}

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}
