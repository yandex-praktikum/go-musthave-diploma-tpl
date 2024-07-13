package client

import (
	"io"
	netHTTP "net/http"
)

type ClientHTTP interface {
	NewRequest(method, url string, body io.Reader) (*netHTTP.Request, error)
	Do(req *netHTTP.Request) (*netHTTP.Response, error)
}
