package domain

import "errors"

var (
	ErrServerInternal              = errors.New("InternalError")               // Ошибка на сервере
	ErrDataFormat                  = errors.New("DataFormatError")             // Неверный формат запроса
	ErrWrongOrderNumber            = errors.New("WrongOrderNumber")            // Неверный номера заказа
	ErrUserNotFound                = errors.New("UserNotFound")                // Пользователь не найден
	ErrLoginIsBusy                 = errors.New("LoginIsBusy")                 // Логин занят
	ErrWrongLoginPassword          = errors.New("WrongLoginPassword")          // Не верный логин/пароль
	ErrAuthDataIncorrect           = errors.New("AuthDataIncorrect")           // Неверный JWT
	ErrNotEnoughPoints             = errors.New("NotEnoughPoints")             // Средств не достаточно
	ErrNotFound                    = errors.New("NoDataFound")                 // Данные не найдены
	ErrOrderNumberAlreadyProcessed = errors.New("OrderNumberAlreadyProcessed") // Данные пользователя уже были приняты в обработку
	ErrDublicateOrderNumber        = errors.New("DublicateOrderNumber")        // Данные уже были приняты в обработку от другого пользователя
	ErrUserIsNotAuthorized         = errors.New("UserIsNotAuthorized")         // Пользователь не авторизован
)
