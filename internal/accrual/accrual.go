package accrual

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"log"
)

type RegisterResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func GetOrderInfo(accrualServerAddress string, orderNumber string) (error, RegisterResponse) {
	var registerResponse RegisterResponse
	req := resty.New().
		SetHostURL(accrualServerAddress).
		R().
		SetHeader("Content-Type", "application/json")

	resp, err := req.Get("/api/user/" + orderNumber)

	if err != nil {
		return err, registerResponse
	}

	if resp.StatusCode() != 200 {
		return err, registerResponse
	}

	if err := json.Unmarshal(resp.Body(), &registerResponse); err != nil {
		return err, registerResponse
	}

	log.Printf("Order: %s, Status: %s, Accrual: %d", registerResponse.Order, registerResponse.Status, registerResponse.Accrual)

	return nil, registerResponse
}
