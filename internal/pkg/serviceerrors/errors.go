package serviceerrors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kdv2001/loyalty/internal/pkg/logger"
)

type AppError struct {
	Msg         string
	Code        int
	Base        error  `json:"-"`
	Description string `json:"description,omitempty"`
}

func NewBadRequest() *AppError {
	return &AppError{"bad request", http.StatusBadRequest, nil, ""}
}

func NewConflict() *AppError {
	return &AppError{"conflict", http.StatusConflict, nil, ""}
}

func NewNotFound() *AppError {
	return &AppError{"not found", http.StatusNotFound, nil, ""}
}

func NewNoContent() *AppError {
	return &AppError{"no content", http.StatusNoContent, nil, ""}
}

func NewTooManyRequests() *AppError {
	return &AppError{"too many requests", http.StatusTooManyRequests, nil, ""}
}

func NewUnauthorized() *AppError {
	return &AppError{"unauthorized", http.StatusUnauthorized, nil, ""}
}

func NewUnprocessableEntity() *AppError {
	return &AppError{"unprocessable entity", http.StatusUnprocessableEntity, nil, ""}
}

func NewAppError(err error) *AppError {
	return &AppError{"internal error", http.StatusInternalServerError, err, ""}
}

func NewPaymentRequired() *AppError {
	return &AppError{"payment required", http.StatusPaymentRequired, nil, ""}
}

func AppErrorFromError(inputError error) *AppError {
	var appErr *AppError
	ok := errors.As(inputError, &appErr)
	if !ok {
		return NewAppError(inputError)
	}
	return appErr
}

func (err *AppError) IsInternalError() bool {
	return err.Code/100 == 5
}

func (err *AppError) Wrap(baseErr error, desc string) *AppError {
	err.Base = baseErr
	err.Description = desc
	return err
}

func (err *AppError) Is(target error) bool {
	targetAppErr := new(AppError)
	ok := errors.As(target, &targetAppErr)
	if !ok {
		return target == err.Base
	}
	return targetAppErr.Code == err.Code && targetAppErr.Msg == err.Msg
}

func (err *AppError) LogServerError(ctx context.Context) *AppError {
	if err.IsInternalError() {
		logger.FromContext(ctx).
			Errorf("[%s] %d %s %v", "", err.Code, err.Description, err.Base)
	}

	return err
}

func (err *AppError) Error() string {
	return err.Msg
}

func (err *AppError) String() string {
	errBuffer, er := json.Marshal(err)
	if er != nil {
		panic(er)
	}
	return string(errBuffer)
}
