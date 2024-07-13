package converter

import (
	"encoding/json"
	"github.com/GTech1256/go-musthave-diploma-tpl/internal/client/accrual/dto"
	"io"
)

func ResponseBodyToOrderDTO(body *io.ReadCloser) (*dto.OrderResponse, error) {
	var order dto.OrderResponse

	decoder := json.NewDecoder(*body)

	err := decoder.Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
