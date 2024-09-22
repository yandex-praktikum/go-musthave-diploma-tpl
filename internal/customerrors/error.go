package customerrors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNotFound          = errors.New("not found")
	ErrNotUser           = errors.New("not user in base")
	ErrNotBonus          = errors.New("not bonus")
	ErrIsTruePassword    = errors.New("password is not true")
	ErrInCorrectMethod   = errors.New("incorrect method")
	ErrAnotherUsersOrder = errors.New("another user`s order")
	ErrOrderIsAlready    = errors.New("the older is already there")
	ErrNotData           = errors.New("not data to answer")
	ErrNotEnoughBonuses  = errors.New("not enough bonuses")
	ErrUserNotFound      = errors.New("user not found")
)

type APIError struct {
	Message string `json:"message"`
}
