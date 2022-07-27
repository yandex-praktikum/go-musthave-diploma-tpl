package externalservice

import (
	"encoding/json"
	"fmt"
	"github.com/botaevg/gophermart/internal/repositories"
	"io"
	"log"
	"net/http"
)

type ExternalService struct {
	storage              repositories.Storage
	accrualSystemAddress string
}

type OrderES struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual uint   `json:"accrual"`
}

func NewES(storage repositories.Storage, accrualSystemAddress string) ExternalService {
	return ExternalService{
		storage:              storage,
		accrualSystemAddress: accrualSystemAddress,
	}
}

func (e ExternalService) AccrualPoints(orderID string) {
	client := http.Client{}

	URL := fmt.Sprintf("%s/api/orders/%s", e.accrualSystemAddress, orderID)
	log.Print(URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Print(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(resp.Status)
	if resp.Status == "200 OK" {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
			return
		}
		defer resp.Body.Close()

		var Order OrderES
		err = json.Unmarshal(respBody, &Order)
		if err != nil {
			log.Print(err)
		}
		log.Print(Order)
	}
}
