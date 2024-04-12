package domain

import "errors"

var (
	ErrServerInternal      = errors.New("InternalError")       // Ошибка на сервере
	ErrDataFormat          = errors.New("DataFormatError")     // Неверный формат запроса
	ErrWrongOrderNumber    = errors.New("WrongOrderNumber")    // Неверный номера заказа
	ErrUserNotFound        = errors.New("UserNotFound")        // Пользователь не найден
	ErrLoginIsBusy         = errors.New("LoginIsBusy")         // Логин занят
	ErrWrongLoginPassword  = errors.New("WrongLoginPassword")  // Не верный логин/пароль
	ErrAuthDataIncorrect   = errors.New("AuthDataIncorrect")   // Неверный JWT
	ErrNotEnoughPoints     = errors.New("NotEnoughPoints")     // Средств не достаточно
	ErrNotFound            = errors.New("NotFoundError")       // Данные не найдены
	ErrDublicateUserData   = errors.New("DublicateUserData")   // Данные пользователя уже были приняты в обработку
	ErrDublicateData       = errors.New("DublicateData")       // Данные уже были приняты в обработку
	ErrUserIsNotAuthorized = errors.New("UserIsNotAuthorized") // Пользователь не авторизован
)
