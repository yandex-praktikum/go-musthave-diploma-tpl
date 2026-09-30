package model

import (
	"errors"
)

var (
	// ErrUserAlreadyExists is returned when registering an occupied username.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrMissingUsername is returned when the username is empty.
	ErrMissingUsername = errors.New("username is required")
	// ErrMissingPassword is returned when the password is empty.
	ErrMissingPassword = errors.New("password is required")
	// ErrUserNotFound is returned when no user with the given username exists.
	ErrUserNotFound = errors.New("user not found")
	// ErrBadPassword is returned when the password fails validation.
	ErrBadPassword = errors.New("bad password")
	// ErrMismatchedPassword is returned when the supplied password does not match the stored hash.
	ErrMismatchedPassword = errors.New("a password and hash do not match")

	ErrMissingOrderNumber = errors.New("order number is required")
	ErrInvalidOrderNumber = errors.New("invalid order number format")

	ErrOrderExistsSameUser  = errors.New("order already loaded by this user")
	ErrOrderExistsOtherUser = errors.New("order already loaded by another user")
)

// User is the registration and login request payload.
type User struct {
	Username string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

type Order struct {
	Username  string `json:"login"`
	Number    string `json:"orderNumber"`
	Processed bool   `json:"processed"`
}

type OrderWriteResult struct {
	Inserted bool
	Owner    string
}

// Validate reports whether username and password are non-empty.
func (user User) Validate() error {
	if user.Username == "" {
		return ErrMissingUsername
	}
	if user.Password == "" {
		return ErrMissingPassword
	}
	return nil
}

func (order Order) Validate() error {
	if order.Number == "" {
		return ErrMissingOrderNumber
	}
	if !checkSum(order.Number) {
		return ErrInvalidOrderNumber
	}
	return nil
}

func checkSum(orderNumber string) bool {
	var sum = 0
	var parity = len(orderNumber) % 2

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
	return int(sum%10) == 0
}
