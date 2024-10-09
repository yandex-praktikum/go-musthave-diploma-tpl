package accrual

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/shopspring/decimal"
	"net/http"
)

type RegisterResponse struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
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

	if resp.StatusCode() != http.StatusOK {
		return registerResponse, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if err := json.Unmarshal(resp.Body(), &registerResponse); err != nil {
		return registerResponse, err
	}

	return registerResponse, nil
}
