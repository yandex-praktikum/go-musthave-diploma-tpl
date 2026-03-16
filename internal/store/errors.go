package store

import "errors"

var (
	ErrRecordNotFound                     = errors.New("record not found")
	ErrOrderNumberAlreadyExistInThisUser  = errors.New("order number already exist in this user")
	ErrOrderNumberAlreadyExistAnotherUser = errors.New("order number already exist another user")
)
