package model

import "errors"

var (
	// ErrMissingOrderNumber is returned when the order number is empty.
	ErrMissingOrderNumber = errors.New("order number is required")

	// ErrInvalidOrderNumber is returned when the order number fails the Luhn check.
	ErrInvalidOrderNumber = errors.New("invalid order number format")

	// ErrOrderExistsSameUser is returned when the order was already loaded by the same user.
	ErrOrderExistsSameUser = errors.New("order already loaded by this user")

	// ErrOrderExistsOtherUser is returned when the order was already loaded by another user.
	ErrOrderExistsOtherUser = errors.New("order already loaded by another user")

	// ErrOrderTableEmpty is returned when the order table contains no rows.
	ErrOrderTableEmpty = errors.New("order table is empty")

	// ErrInsufficientFunds is returned when the account balance cannot cover the withdrawal.
	ErrInsufficientFunds = errors.New("insufficient funds in the account")
)
