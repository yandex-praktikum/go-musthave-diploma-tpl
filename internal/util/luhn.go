package util

import "strconv"

func IsValid(order string) bool {
	number, err := strconv.Atoi(order)
	if err != nil {
		return false
	}

	var sum int

	for i := 0; number > 0; i++ {
		last := number % 10
		if i%2 != 0 {
			last *= 2
			if last > 9 {
				last -= 9
			}
		}
		sum += last
		number /= 10
	}
	return sum%10 == 0
}
