package model

import "errors"

// Order-related errors
var (
	// ErrMissingOrderNumber is returned when the order number is empty.
	ErrMissingOrderNumber = errors.New("order number is required")

	// ErrInvalidOrderNumber is returned when the order number fails the Luhn check.
	ErrInvalidOrderNumber = errors.New("invalid order number format")

	// ErrOrderExistsSameUser is returned when the order was already loaded by the same user.
	ErrOrderExistsSameUser = errors.New("order already loaded by this user")

	// ErrOrderExistsOtherUser is returned when the order was already loaded by another user.
	ErrOrderExistsOtherUser = errors.New("order already loaded by another user")
)

// User-related errors
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
)

// Withdrawal-related errors
var (
	// ErrInsufficientFunds is returned when the account balance cannot cover the withdrawal.
	ErrInsufficientFunds = errors.New("insufficient funds in the account")
)
