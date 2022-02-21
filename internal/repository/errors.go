package repository

import "errors"

var (
	ErrInt         = errors.New("error: internal db error")
	ErrOrdOverLap  = errors.New("error: order already exist")
	ErrLoginConfl  = errors.New("error: login already exist")
	ErrOrdUsrConfl = errors.New("error: order was added by other customer")
	ErrUsrUncor    = errors.New("error: username or password is not correct")
)
