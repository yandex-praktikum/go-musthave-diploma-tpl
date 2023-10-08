package httperrors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

const (
	ErrBadRequest          = "Bad request"
	ErrStatusConflict      = "Status conflict"
	ErrNotFound            = "Not Found"
	ErrUnauthorized        = "Unauthorized"
	ErrRequestTimeout      = "Request Timeout"
	ErrInvalidPassword     = "Invalid password"
	ErrInvalidField        = "Invalid field"
	ErrInternalServerError = "Internal Server Error"
)

// RestErr Rest error interface
type RestErr interface {
	Status() int
	Error() string
	Causes() interface{}
	ErrBody() RestError
}

// RestError Rest error struct
type RestError struct {
	ErrStatus  int         `json:"status,omitempty"`
	ErrError   string      `json:"error,omitempty"`
	ErrMessage interface{} `json:"message,omitempty"`
	Timestamp  time.Time   `json:"timestamp,omitempty"`
}

// ErrBody Error body
func (e RestError) ErrBody() RestError {
	return e
}

// Error  Error() interface method
func (e RestError) Error() string {
	return fmt.Sprintf("status: %d - errors: %s - causes: %v", e.ErrStatus, e.ErrError, e.ErrMessage)
}

// Status Error status
func (e RestError) Status() int {
	return e.ErrStatus
}

// Causes RestError Causes
func (e RestError) Causes() interface{} {
	return e.ErrMessage
}

// NewRestError New Rest Error
func NewRestError(status int, err string, causes interface{}, debug bool) RestErr {
	restError := RestError{
		ErrStatus: status,
		ErrError:  err,
		Timestamp: time.Now().UTC(),
	}
	if debug {
		restError.ErrMessage = causes
	}
	return restError
}

// ParseErrors Parser of error string messages returns RestError
func ParseErrors(err error, debug bool) RestErr {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return NewRestError(http.StatusNotFound, ErrNotFound, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "unauthorized"):
		return NewRestError(http.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "signature is invalid"):
		return NewRestError(http.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "conflict"):
		return parseSqlErrors(err, debug)
	case strings.Contains(strings.ToLower(err.Error()), "unique constraint"):
		return parseSqlErrors(err, debug)
	case strings.Contains(strings.ToLower(err.Error()), "field validation"):
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return NewRestError(http.StatusBadRequest, ErrBadRequest, validationErrors.Error(), debug)
		}
		return parseValidatorError(err, debug)
	default:
		if restErr, ok := err.(*RestError); ok {
			return restErr
		}
		return NewRestError(http.StatusInternalServerError, ErrInternalServerError, err, debug)
	}
}

func parseSqlErrors(err error, debug bool) RestErr {
	return NewRestError(http.StatusConflict, ErrStatusConflict, err, debug)
}

func parseValidatorError(err error, debug bool) RestErr {

	if strings.Contains(err.Error(), "Password") {
		return NewRestError(http.StatusBadRequest, ErrInvalidPassword, err, debug)
	}

	return NewRestError(http.StatusBadRequest, ErrInvalidField, err, debug)
}
