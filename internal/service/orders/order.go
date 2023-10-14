package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (s Service) CreateOrder(usedID string, order []byte) int {
	status := models.NEW
	orderID := string(order)
	valid := ValidateLuhnNumber(orderID)
	fmt.Print(valid)
	if !valid {
		return models.WrongOrderFormat
	}

	model := models.SaveOrder{
		UserID:  usedID,
		OrderID: orderID,
		Status:  status,
	}

	s.accrual <- model.OrderID

	res := s.orders.SaveUserOrder(model)

	return res
}

func (s Service) CheckOrder(number string) {

	url := fmt.Sprintf("%S/api/orders/%s", s.conf.Accrual, number)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
	}
	request.Header.Set("Content-Length", "0")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
	}
	defer response.Body.Close()

	for {
		if response.StatusCode != http.StatusNoContent {
			_ = fmt.Errorf("request failed: %s", response.Status)
		} else if response.StatusCode == http.StatusTooManyRequests {
			_ = fmt.Errorf("too many requests: %s", response.Status)
			retryAfter, err := strconv.Atoi(response.Header.Get("Retry-After"))
			if err != nil {
				break
			}
			time.Sleep(time.Duration(retryAfter))
		} else if response.StatusCode == http.StatusOK {
			break
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("error reading response body: %s", err)
	}

	orderResponse := models.LoyaltySystem{}
	err = json.Unmarshal(body, &orderResponse)
	if err != nil {
		fmt.Printf("error decoding response: %s", err)
	}

	s.orderStatus <- orderResponse

	fmt.Print("order res :", orderResponse)

	return

}

func (s Service) GetOrders() {
	//TODO implement me
	return
}

func (s Service) ChangeOrderStatus(ctx context.Context) (context.Context, func()) {
	fmt.Print("change order started")

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case data := <-s.orderStatus:
			fmt.Print(data)
			s.orders.ChangeOrderStatus(data)
		}
	}
}

func (s Service) AccrualOrderStatus(ctx context.Context) (context.Context, func()) {
	fmt.Print("accrual started")
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case data := <-s.accrual:
			fmt.Print(data)
			s.CheckOrder(data)
		}
	}
}
