package models

import (
	"errors"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrExists          = errors.New("already exists")
	ErrNoEnoughBalance = errors.New("not enough money on balance")
	ErrOrderLoaded     = errors.New("already loaded order number")
	ErrWrongUser       = errors.New("already loaded by another user")
)
