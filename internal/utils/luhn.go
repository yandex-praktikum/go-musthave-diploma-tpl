package utils

func LuhnCheck(number string) bool {
	var sum int
	for i, r := range reverse(number) {
		digit := int(r - '0')
		if i%2 == 1 {
			digit *= 2
		}
		if digit > 9 {
			digit -= 9
		}
		sum += digit
	}
	return sum%10 == 0
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
