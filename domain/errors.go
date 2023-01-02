package domain

import "errors"

var (
	ErrLoginIsBusy         = errors.New("login is busy")
	ErrLoadedByThisUser    = errors.New("loaded by this user")
	ErrLoadedByAnotherUser = errors.New("loaded by another user")
)
