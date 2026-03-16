package http

import "net/http"

type httpClientMock struct {
	err      error
	response http.Response
}

func (c *httpClientMock) Do(_ *http.Request) (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}

	return &c.response, nil
}
