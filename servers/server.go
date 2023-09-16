package servers

import "net/http"

type (
	RequestType int
)

const (
	Post RequestType = 1 << iota
	Get
)

type Server interface {
	Start(addr string) error
	AddHandler(t RequestType, pattern string, handlerFn http.HandlerFunc)
}
