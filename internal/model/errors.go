package model

import "fmt"

var (
	ErrorUnverifiedAccount = fmt.Errorf("login unverified")
	ErrorUserAlreadyExists = fmt.Errorf("user already exists")
	ErrorNoUser            = fmt.Errorf("user doesn't exists")

	ErrorWrongOrders        = fmt.Errorf("invalid order")
	ErrorOrderAlreadyExists = fmt.Errorf("order already exists")
	ErrorOrderConflict      = fmt.Errorf("order already register")

	ErrorInsufficientBalance = fmt.Errorf("insufficient balance")
)
