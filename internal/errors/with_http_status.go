package errors

import (
	"errors"
	"net/http"
)

type ErrorWithHTTPStatus struct {
	Message    string
	StatusCode int
}

func (e *ErrorWithHTTPStatus) Error() string {
	return e.Message
}

func NewErrorWithHTTPStatus(message string, statusCode int) error {
	return &ErrorWithHTTPStatus{
		Message:    message,
		StatusCode: statusCode,
	}
}

func GetMessageAndStatusCode(err error) (string, int) {
	var ewhs *ErrorWithHTTPStatus
	if errors.As(err, &ewhs) {
		return ewhs.Message, ewhs.StatusCode
	} else {
		return err.Error(), http.StatusInternalServerError
	}
}
