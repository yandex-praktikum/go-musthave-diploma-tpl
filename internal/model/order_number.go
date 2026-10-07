package model

// ValidateOrderNumber reports whether the order number is non-empty and passes the Luhn check.
func ValidateOrderNumber(number string) error {
	if number == "" {
		return ErrMissingOrderNumber
	}
	if !luhnCheck(number) {
		return ErrInvalidOrderNumber
	}
	return nil
}

// luhnCheck reports whether orderNumber passes the Luhn checksum algorithm.
func luhnCheck(orderNumber string) bool {
	if len(orderNumber) == 0 {
		return false
	}
	sum := 0
	parity := len(orderNumber) % 2

	for i := 0; i < len(orderNumber); i++ {
		if orderNumber[i] < '0' || orderNumber[i] > '9' {
			return false
		}
		digit := int(orderNumber[i] - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
