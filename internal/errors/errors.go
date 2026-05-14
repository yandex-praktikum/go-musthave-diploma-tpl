// Пакет для кастомных ошибок
package errors

import "net/http"

type AppError struct {
	Code int
	Msg  string
	Log  string
}

func NewBadRequestError(message, log string) *AppError {
	return &AppError{
		Code: http.StatusBadRequest,
		Msg:  message,
		Log:  log,
	}
}

func NewUnauthorizedError(log string) *AppError {
	return &AppError{
		Code: http.StatusUnauthorized,
		Msg:  http.StatusText(http.StatusUnauthorized),
		Log:  log,
	}
}

func NewPaymentRequiredError(log string) *AppError {
	return &AppError{
		Code: http.StatusPaymentRequired,
		Msg:  http.StatusText(http.StatusPaymentRequired),
		Log:  log,
	}
}

func NewNotFoundError(message, log string) *AppError {
	return &AppError{
		Code: http.StatusNotFound,
		Msg:  message,
		Log:  log,
	}
}

func NewConflictError(log string) *AppError {
	return &AppError{
		Code: http.StatusConflict,
		Msg:  http.StatusText(http.StatusConflict),
		Log:  log,
	}
}

func NewUnprocessableEntityError(log string) *AppError {
	return &AppError{
		Code: http.StatusUnprocessableEntity,
		Msg:  http.StatusText(http.StatusUnprocessableEntity),
		Log:  log,
	}
}

func NewInternalServerError(log string) *AppError {
	return &AppError{
		Code: http.StatusInternalServerError,
		Msg:  http.StatusText(http.StatusInternalServerError),
		Log:  log,
	}
}
