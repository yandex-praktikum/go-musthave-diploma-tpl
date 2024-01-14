package externalservice

import (
	"encoding/json"
	"fmt"
	"github.com/botaevg/gophermart/internal/models"
	"github.com/botaevg/gophermart/internal/service"
	"io"
	"log"
	"net/http"
)

type ExternalService struct {
	gophermart           service.Gophermart
	asyncExecution       chan string
	accrualSystemAddress string
}

func NewES(accrualSystemAddress string, gophermart service.Gophermart, asyncExecution chan string) ExternalService {
	return ExternalService{
		accrualSystemAddress: accrualSystemAddress,
		gophermart:           gophermart,
		asyncExecution:       asyncExecution,
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

		var Order models.OrderES
		err = json.Unmarshal(respBody, &Order)
		if err != nil {
			log.Print(err)
		}
		log.Print(Order)

		e.gophermart.UpdateOrders(Order)

		if Order.Status == "PROCESSED" {
			e.gophermart.AccrualRequest(Order)
		}

		if Order.Status == "REGISTERED" || Order.Status == "PROCESSING" {
			e.asyncExecution <- orderID
			return
		}

	} else {
		e.asyncExecution <- orderID
	}
}
