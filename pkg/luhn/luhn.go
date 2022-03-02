package luhn

func Validate(number string) bool {
	var sum int
	sum = 0
	len := len(number)
	isSecond := false
	for i := len - 1; i >= 0; i-- {
		d := int(number[i]) - '0'
		if isSecond {
			d = d * 2
		}
		sum = sum + d/10
		sum = sum + d%10
		isSecond = !isSecond
	}

	return (sum%10 == 0)
}
