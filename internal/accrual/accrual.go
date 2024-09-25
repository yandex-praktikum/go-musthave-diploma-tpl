package accrual

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
)

type RegisterResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func GetOrderInfo(accrualServerAddress string, orderNumber string) (RegisterResponse, error) {
	var registerResponse RegisterResponse
	req := resty.New().
		SetBaseURL(accrualServerAddress).
		R().
		SetHeader("Content-Type", "application/json")

	resp, err := req.Get("/api/orders/" + orderNumber)

	if err != nil {
		return registerResponse, err
	}

	if resp.StatusCode() != 200 {
		return registerResponse, err
	}

	if err := json.Unmarshal(resp.Body(), &registerResponse); err != nil {
		return registerResponse, err
	}

	return registerResponse, nil
}
