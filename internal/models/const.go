package models

type IDKey string

const UserIDKey IDKey = "userID"
const EXIST = "exist"

const (
	REGISTERED = "REGISTERED"
	INVALID    = "INVALID"
	PROCESSING = "PROCESSING"
	PROCESSED  = "PROCESSED"
	NEW        = "NEW"
)

const (
	COOKIE    = "go_session"
	APPJSON   = "application/json"
	TXTHTML   = "text/html"
	TEXTPlAIN = "text/plain"
	CONTENT   = "Content-Type"
)

const (
	ErrInvalidBody = "invalid request body"
)

const (
	DownloadedByUser        = 200
	Accepted                = 202
	WrongRequestFormat      = 400
	UserUnauthorized        = 401
	DownloadedByAnotherUser = 409
	WrongOrderFormat        = 422
	InternalServerError     = 500
)
