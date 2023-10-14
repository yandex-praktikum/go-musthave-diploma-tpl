package order

import "net/http"

const (
	AlreadyDownloaded   = http.StatusOK
	Accepted            = http.StatusAccepted
	WrongFormat         = http.StatusBadRequest
	Unauthorized        = http.StatusUnauthorized
	NoPermissions       = http.StatusConflict
	WrongOrderFormat    = http.StatusUnprocessableEntity
	InternalServerError = http.StatusInternalServerError
)
