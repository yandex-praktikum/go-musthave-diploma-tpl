package entities

import (
	"encoding/json"
	"io"
	"net/http"
)

type GetOrderClientResponseBody struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

var _ ResponseParser = &GetOrderClientResponseBody{}

func (b *GetOrderClientResponseBody) ParseFromResponse(res *http.Response) error {
	rawResponseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if !json.Valid(rawResponseBody) {
		return err
	}

	if err = json.Unmarshal(rawResponseBody, b); err != nil {
		return err
	}
	return nil
}
