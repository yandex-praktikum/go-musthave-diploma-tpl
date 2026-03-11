package service

import "fmt"

// ErrInvalidCredentials — неверная пара логин/пароль.
type ErrInvalidCredentials struct {
	Login string
}

func (e *ErrInvalidCredentials) Error() string {
	return fmt.Sprintf("invalid credentials for login %q", e.Login)
}

// ErrValidation — ошибка валидации (пустой логин/пароль, неверный номер заказа и т.д.).
type ErrValidation struct {
	Msg string
}

func (e *ErrValidation) Error() string {
	return e.Msg
}

// ErrOrderOwnedByOther — номер заказа уже загружен другим пользователем (409).
type ErrOrderOwnedByOther struct {
	Number string
}

func (e *ErrOrderOwnedByOther) Error() string {
	return fmt.Sprintf("order %q already uploaded by another user", e.Number)
}
