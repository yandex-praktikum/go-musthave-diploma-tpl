package entities

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WithdrawRequestData struct {
	OrderID string  `json:"order"`
	Sum     float64 `json:"sum"`
}

var _ RequestParser = &WithdrawRequestData{}

func (b *WithdrawRequestData) ParseFromRequest(req *http.Request) error {
	rawRequestBody, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if !json.Valid(rawRequestBody) {
		return fmt.Errorf("invalid json")
	}

	if err = json.Unmarshal(rawRequestBody, b); err != nil {
		return err
	}

	return nil
}
