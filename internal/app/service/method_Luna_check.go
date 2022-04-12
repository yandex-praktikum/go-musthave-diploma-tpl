package service

// LuhnCheck verify the provided order number (string format) and returns true if it's correct
func LuhnCheck(number string) bool {
	sum := number[len(number)-1] - '0'

	for i := len(number) - 2; i >= 0; i-- {
		n := number[i] - '0'
		if i%2 == len(number)%2 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}

	return sum%10 == 0
}
