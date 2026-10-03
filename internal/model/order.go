package model

// Status is the order processing state in the accrual flow.
type Status string

// Order lifecycle statuses
const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
	StatusNew        = "NEW"
)

// Order is a loyalty order bound to a user; it is the internal system contract.
type Order struct {
	Username string `json:"login"`
	Number   string `json:"order"`
	Status   string `json:"status,omitempty"`
	Accrual  int    `json:"accrual,omitempty"`
}

// OrderWriteResult reports whether an order was inserted and, if not, who owns it.
type OrderWriteResult struct {
	Inserted bool
	Owner    string
}

// Validate reports whether the order number is non-empty and passes the Luhn check.
func (order Order) Validate() error {
	if order.Number == "" {
		return ErrMissingOrderNumber
	}
	if !luhnCheck(order.Number) {
		return ErrInvalidOrderNumber
	}
	return nil
}

// luhnCheck reports whether orderNumber passes the Luhn checksum algorithm.
func luhnCheck(orderNumber string) bool {
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
