package auth

import "net/http"

const (
	SuccessAuth         = http.StatusOK
	WrongRequestFormat  = http.StatusBadRequest
	LoginIsTaken        = http.StatusConflict
	InternalServerError = http.StatusInternalServerError
	WrongLoginPassword  = http.StatusUnauthorized
)
