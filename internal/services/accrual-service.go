package services

import (
	"encoding/json"
	"fmt"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

type AccrualService struct {
	accrualBaseURL      string
	httpClient          *http.Client
	orderChan           chan entities.Order
	failedOrderChan     chan entities.Order
	accountRepository   repositories.AccountRepositoryInterface
	operationRepository repositories.OperationRepositoryInterface
	orderRepository     repositories.OrderRepositoryInterface
}

func NewAccrualService(
	accrualBaseURL string,
	accountRepository repositories.AccountRepositoryInterface,
	operationRepository repositories.OperationRepositoryInterface,
	orderRepository repositories.OrderRepositoryInterface,
) *AccrualService {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	orderChan := make(chan entities.Order, 1000)
	failedOrderChan := make(chan entities.Order, 1000)

	instance := &AccrualService{
		accrualBaseURL,
		httpClient,
		orderChan,
		failedOrderChan,
		accountRepository,
		operationRepository,
		orderRepository,
	}

	return instance
}

func (ac *AccrualService) SendOrderToQueue(order entities.Order) {
	ac.orderChan <- order
}

func (ac *AccrualService) ProcessOrders(e *echo.Echo) {
	for order := range ac.orderChan {
		ac.processOrder(e, order)
	}
}

func (ac *AccrualService) ProcessFailedOrders(e *echo.Echo) {
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		orders, err := ac.orderRepository.GetOrdersForProcess()
		if err != nil {
			e.Logger.Error(err.Error())
			continue
		}
		for _, order := range orders {
			ac.processOrder(e, *order)
		}

	}
}

func (ac *AccrualService) processOrders(e *echo.Echo, orderChan chan entities.Order) {
}

func (ac *AccrualService) processOrder(e *echo.Echo, order entities.Order) {
	accrualOrder, err := ac.fetchOrder(e, order.Number)
	if err != nil {
		return
	}

	switch accrualOrder.Status {
	case entities.OrderStatusNew:
	case entities.OrderStatusProcessing:
		err = ac.orderRepository.UpdateOrderByAccrualOrder(accrualOrder)
		if err != nil {
			return
		}
		ac.orderChan <- order
	case entities.OrderStatusProcessed:
		err = ac.orderRepository.UpdateOrderByAccrualOrder(accrualOrder)
		if err != nil {
			return
		}

		bonusAccount, err := ac.accountRepository.FindByUserID(order.UserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			return
		}
		err = ac.operationRepository.CreateAccrual(bonusAccount.ID, accrualOrder.Order, accrualOrder.Accrual)
		if err != nil {
			return
		}
	case entities.OrderStatusInvalid:
		_ = ac.orderRepository.UpdateOrderByAccrualOrder(accrualOrder)
		return
	}
}

func (ac *AccrualService) fetchOrder(e *echo.Echo, orderNumber string) (*models.AccrualOrderResponse, error) {
	res := &models.AccrualOrderResponse{
		Order:  orderNumber,
		Status: entities.OrderStatusProcessing,
	}

	url := fmt.Sprintf("%s/api/orders/%s", ac.accrualBaseURL, orderNumber)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		e.Logger.Error(err)
		return res, errors.New("cannot create request: " + err.Error())
	}

	response, err := ac.httpClient.Do(request)
	if err != nil {
		e.Logger.Error(err)
		return res, errors.New("cannot get order: " + err.Error())
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			e.Logger.Error(err)
			return res, errors.New("cannot read response: " + err.Error())
		}

		err = json.Unmarshal(body, &res)
		if err != nil {
			e.Logger.Error(err)
			return res, errors.New("cannot json unmarshal: " + err.Error())
		}

		return res, nil
	case http.StatusNoContent:
		res.Status = entities.OrderStatusProcessing
		return res, nil
	case http.StatusTooManyRequests:
		e.Logger.Info("many requests ", orderNumber)
		return res, errors.New("response too many request")
	case http.StatusInternalServerError:
		e.Logger.Info("internal server error", orderNumber)
		return res, errors.New("response internal server error")
	default:
		e.Logger.Info("response unknown status: "+strconv.Itoa(response.StatusCode), orderNumber)
		return res, errors.New("response unknown status: " + strconv.Itoa(response.StatusCode))
	}
}
