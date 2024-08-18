package utils

import "strconv"

func CheckLuhnAlg(numberStr string) bool {
	sum := 0
	nDigits := len(numberStr)
	parity := nDigits % 2
	for index, digitStr := range numberStr {
		digitInt, err := strconv.Atoi(string(digitStr))
		if err != nil {
			return false
		}
		if index%2 == parity {
			digitInt *= 2
			if digitInt > 9 {
				digitInt -= 9
			}
		}
		sum += digitInt
	}
	return sum%10 == 0
}
