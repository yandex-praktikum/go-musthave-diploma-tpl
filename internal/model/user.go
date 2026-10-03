package model

import (
	"errors"
)

var (
	// ErrUserAlreadyExists is returned when registering an occupied username.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrMissingUsername is returned when the username is empty.
	ErrMissingUsername = errors.New("username is required")
	// ErrMissingPassword is returned when the password is empty.
	ErrMissingPassword = errors.New("password is required")
	// ErrUserNotFound is returned when no user with the given username exists.
	ErrUserNotFound = errors.New("user not found")
	// ErrBadPassword is returned when the password fails validation.
	ErrBadPassword = errors.New("bad password")
	// ErrMismatchedPassword is returned when the supplied password does not match the stored hash.
	ErrMismatchedPassword = errors.New("a password and hash do not match")
)

// User is the registration and login request payload.
type User struct {
	Username string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

// Validate reports whether username and password are non-empty.
func (user User) Validate() error {
	if user.Username == "" {
		return ErrMissingUsername
	}
	if user.Password == "" {
		return ErrMissingPassword
	}
	return nil
}
