package clients

import "net/http"

type Client interface {
	DoGet(path string) (*http.Response, error)
}
