package custom_errors

import (
	"errors"
	"net/http"
)

type ErrorWithHttpStatus struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *ErrorWithHttpStatus) Error() string {
	return e.Message
}

func NewErrorWithHttpStatus(message string, statusCode int) error {
	return &ErrorWithHttpStatus{
		Message:    message,
		StatusCode: statusCode,
	}
}

func GetMessageAndStatusCode(err error) (string, int) {
	var ewhs *ErrorWithHttpStatus
	if errors.As(err, &ewhs) {
		return ewhs.Message, ewhs.StatusCode
	} else {
		return err.Error(), http.StatusInternalServerError
	}
}
