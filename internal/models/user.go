package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	MinPasswordLength = 8
	MinLoginLength    = 4
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrWrongPassword       = errors.New("wrong password")
	ErrUserExists          = errors.New("user exists")
	ErrWrongLoginLength    = fmt.Errorf("login length should be at least %d characters", MinLoginLength)
	ErrWrongPasswordLength = fmt.Errorf("password length should be at least %d characters", MinPasswordLength)
)

type (
	/*
		Объект пользователя системы
	*/
	User struct {
		ID        string    `json:"id"`
		Login     string    `json:"login"`
		Password  string    `json:"password"`
		Hash      string    `json:"-"`
		Version   int       `json:"version"`
		CreatedAt time.Time `json:"created,omitempty"`
	}

	/*
		Интерфейс объекта пользователя системы
	*/
	UserInterface interface {
		ToJSON() ([]byte, error)
		Validate() error
	}
)

/*
Конструктор объекта пользователя.
*/
func NewUser(login, password, hash string, version int) *User {
	return &User{Login: login, Password: password, Hash: hash, Version: version}
}

/*
Валидация длины логина и пароля
*/
func (m *User) Validate() error {
	if len(m.Login) < MinLoginLength {
		return ErrWrongLoginLength
	} else if len(m.Password) < MinPasswordLength {
		return ErrWrongPasswordLength
	}
	return nil
}

/*
Сериализует объект пользователя в строку JSON
Может вернуть ошибку, коли та случится.
*/
func (m *User) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
