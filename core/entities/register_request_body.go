package entities

import (
	"encoding/json"
	"io"
	"net/http"
)

type RegisterRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var _ RequestParser = &RegisterRequestBody{}

func (b *RegisterRequestBody) ParseFromRequest(req *http.Request) error {
	rawRequestBody, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if !json.Valid(rawRequestBody) {
		return err
	}

	if err = json.Unmarshal(rawRequestBody, b); err != nil {
		return err
	}
	return nil
}
