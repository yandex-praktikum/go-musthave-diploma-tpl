package entities

import (
	"io"
	"net/http"
)

type AddOrderRequestBody struct {
	OrderID string `json:"order_id"`
}

var _ RequestParser = &AddOrderRequestBody{}

func (b *AddOrderRequestBody) ParseFromRequest(req *http.Request) error {
	rawRequestBody, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	b.OrderID = string(rawRequestBody)

	return nil
}
