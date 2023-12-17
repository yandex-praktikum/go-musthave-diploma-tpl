package utils

import (
	"errors"
	"strconv"
)

func ValidateOrder(number string) error {
	return ValidateMoon(number)
}

func ValidateMoon(number string) error {

	result := make([]int64, 0)

	for i, r := range number {
		val, err := strconv.ParseInt(string(r), 10, 64)

		if err != nil {
			return err
		}
		if i%2 == 0 {

			m := val * 2
			if m > 9 {
				result = append(result, m-9)
			} else {
				result = append(result, m)
			}
		} else {
			result = append(result, val)
		}
	}

	resultSum := sum(result)

	if resultSum%10 == 0 {
		return nil
	}

	return ErrInvalidNumber
}

var (
	ErrInvalidNumber = errors.New("invalid number")
)

func sum(sl []int64) int64 {
	var sum int64 = 0
	for _, val := range sl {
		sum += val
	}
	return sum
}
