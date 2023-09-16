package entities

import "net/http"

type RequestParser interface {
	ParseFromRequest(req *http.Request) error
}

type ResponseParser interface {
	ParseFromResponse(req *http.Response) error
}
