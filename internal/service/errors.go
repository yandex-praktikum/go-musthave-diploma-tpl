package service

import "errors"

var (
	ErrInt      = errors.New("error: internal error")
	ErrNotValid = errors.New("error: not valid number")
	ErrNoMoney  = errors.New("error: not enough bonuses")
)
