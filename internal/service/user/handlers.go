package user

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
)

// ------------ SignUp ------------

// signUpInput DTO
type signUpInput struct {
	Body struct {
		Login    string `json:"login" minLength:"3" maxLength:"32" pattern:"^[a-zA-Z0-9_]+$" doc:"Логин пользователя"`
		Password string `json:"password" minLength:"12" maxLength:"72" pattern:"^[A-Za-z\\d@$!%*?&]{12,72}$" doc:"Пароль пользователя"`
	}
	UserAgent     string `header:"User-Agent" maxLength:"255" doc:"User-Agent браузера пользователя"`
	XForwardedFor string `header:"X-Forwarded-For" maxLength:"255" doc:"X-Forwarded-For заголовок HTTP Proxy"`
}

// signUpOutput DTO
type signUpOutput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
	SetCookie     string `header:"Set-Cookie" doc:"Установка сессионной куки"`
	Body          struct {
		Code   int    `json:"code" example:"200"  doc:"Код выполнения"`
		Status string `json:"status" example:"Success" doc:"Статус операции"`
		Token  string `json:"token" doc:"Короткоживущий JWT токен"`
	}
}

// signUpHandler возвращает HUMA хэндлер регистрации нового пользователя
func (u *User) signUpHandler() func(ctx context.Context, in *signUpInput) (*signUpOutput, error) {

	return func(ctx context.Context, in *signUpInput) (*signUpOutput, error) {
		res, err := u.SignUp(ctx, in)
		if err != nil {
			if errors.Is(err, ErrUserAlreadyExists) {
				return nil, huma.Error409Conflict("Логин уже занят")
			}

			return nil, huma.Error500InternalServerError("Внутренняя ошибка сервера")
		}

		return res, nil
	}

}

// ------------ SignIn ------------

// signInInput DTO
type signInInput struct {
	Body struct {
		Login    string `json:"login" minLength:"3" maxLength:"32" pattern:"^[a-zA-Z0-9_]+$" doc:"Логин пользователя"`
		Password string `json:"password" minLength:"12" maxLength:"72" pattern:"^[A-Za-z\\d@$!%*?&]{12,72}$" doc:"Пароль пользователя"`
	}
	UserAgent     string `header:"User-Agent" maxLength:"255" doc:"User-Agent браузера пользователя"`
	XForwardedFor string `header:"X-Forwarded-For" maxLength:"255" doc:"X-Forwarded-For заголовок HTTP Proxy"`
}

// signInOutput DTO
type signInOutput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
	SetCookie     string `header:"Set-Cookie" doc:"Установка сессионной куки"`
	Body          struct {
		Code   int    `json:"code" example:"200"  doc:"Код выполнения"`
		Status string `json:"status" example:"Success" doc:"Статус операции"`
		Token  string `json:"token" doc:"Короткоживущий JWT токен"`
	}
}

// signInHandler возвращает HUMA хэндлер залогина пользователя
func (u *User) signInHandler() func(ctx context.Context, in *signInInput) (*signInOutput, error) {

	return func(ctx context.Context, in *signInInput) (*signInOutput, error) {
		res, err := u.SignIn(ctx, in)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidPassword) {
				return nil, huma.Error401Unauthorized("Неверные логин или пароль")
			}

			return nil, huma.Error500InternalServerError("Внутренняя ошибка сервера")
		}

		return res, nil
	}

}
