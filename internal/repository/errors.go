package repository

import (
	"fmt"
)

var (
	ErrorLoginAlreadyExists = fmt.Errorf("login already exists")
	ErrorNoUser             = fmt.Errorf("user not exists")

	ErrorOrderAlreadyExists = fmt.Errorf("order already exists")
	ErrorOrderConflict      = fmt.Errorf("order already register")

	ErrorInsufficientBalance = fmt.Errorf("insufficient balance")
)
