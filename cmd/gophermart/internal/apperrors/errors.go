package apperrors

import "errors"

var (
	ErrUserExists              = errors.New("user already exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrAuth                    = errors.New("invalid login or password")
	ErrOrderExists             = errors.New("order already exists")
	ErrOrderOwnedByAnotherUser = errors.New("order uploaded by another user")
	ErrNoMoney                 = errors.New("no money")
)
