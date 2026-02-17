package utils

import "fmt"

func GetOrderInfoKey(userId, orderNumber string) string {
	return fmt.Sprintf("%v%v", GetOrderInfoKeyPrefix(userId), orderNumber)
}

func GetOrderInfoKeyPrefix(userId string) string {
	return fmt.Sprintf("%v_", userId)
}

func ValidLuhn(number string) bool {
	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		d := number[i] - '0'
		if d > 9 {
			return false
		}

		n := int(d)

		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}

		sum += n
		double = !double
	}

	return sum%10 == 0
}
