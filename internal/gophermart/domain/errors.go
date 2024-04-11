package domain

import "errors"

var (
	ErrServerInternal      = errors.New("InternalError")       // Ошибка на сервере
	ErrDataFormat          = errors.New("DataFormatError")     // Ошибка в данных
	ErrDataIsBusy          = errors.New("DataIsAlreadyTaken")  // Данные уже были использованы(логин занят/номер заказа уже был загружен другим пользователем)
	ErrAuthDataIncorrect   = errors.New("AuthDataIsIncorrect") // Неверная пара логин/пароль
	ErrUserIsNotAuthorized = errors.New("UserIsNotAuthorized") // Пользователь не авторизован
	ErrDublicateUserData   = errors.New("DublicateUserData")   // Данные пользователя уже были приняты в обработку
	ErrNotEnoughPoints     = errors.New("NotEnoughPoints")     // Средств не достаточно
	ErrNotFound            = errors.New("NotFoundError")
	ErrDBConnection        = errors.New("DatabaseConnectionError")
)
