package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eac0de/gophermart/internal/errors"
	"github.com/eac0de/gophermart/internal/models"
	"github.com/eac0de/gophermart/pkg/utils"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type OrderStore interface {
	InsertOrder(ctx context.Context, order *models.Order) error
	UpdateOrder(ctx context.Context, order *models.Order) error
	SelectUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
	SelectOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	SelectOrdersForProccesing(ctx context.Context) ([]*models.Order, error)
	UpdateUser(ctx context.Context, user *models.User) error
	SelectUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type OrderService struct {
	client               *resty.Client
	orderStore           OrderStore
	accrualSystemAddress string
}

func NewOrderService(orderStore OrderStore, client *resty.Client, accrualSystemAddress string) *OrderService {
	return &OrderService{
		client:               client,
		orderStore:           orderStore,
		accrualSystemAddress: accrualSystemAddress,
	}
}

func (os *OrderService) AddOrder(ctx context.Context, number string, userID uuid.UUID) (*models.Order, error) {
	if number == "" {
		return nil, errors.NewErrorWithHTTPStatus("order number cannot be empty", http.StatusBadRequest)
	}
	if !utils.CheckLuhnAlg(number) {
		return nil, errors.NewErrorWithHTTPStatus("order number did not pass the Luhn algorithm check", http.StatusUnprocessableEntity)
	}
	order, _ := os.orderStore.SelectOrderByNumber(ctx, number)
	if order != nil {
		if order.UserID == userID {
			return nil, errors.NewErrorWithHTTPStatus("order number has already been uploaded by you", http.StatusOK)
		}
		return nil, errors.NewErrorWithHTTPStatus("order number has already been uploaded by another user", http.StatusConflict)
	}
	order = &models.Order{
		ID:         uuid.New(),
		Number:     number,
		Status:     models.OrderStatusNew,
		UserID:     userID,
		UploadedAt: time.Now(),
	}
	err := os.orderStore.InsertOrder(ctx, order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (os *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	return os.orderStore.SelectUserOrders(ctx, userID)
}

func (os *OrderService) StartProcessingOrders(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			orders, err := os.orderStore.SelectOrdersForProccesing(ctx)
			if err != nil {
				log.Printf("fetching orders error: %v\n", err)
				continue
			}
			if len(orders) == 0 {
				continue
			}
			for _, order := range orders {
				err := os.SendOrderForCalculation(ctx, order)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}
}

func (os *OrderService) SendOrderForCalculation(ctx context.Context, order *models.Order) error {
	url := fmt.Sprintf("%s/api/orders/%v", os.accrualSystemAddress, order.Number)
	request := os.client.
		R().
		SetHeader("Accept-Encoding", "gzip")
	resp, err := request.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusNoContent {
		order.Status = models.OrderStatusInvalid
		return os.orderStore.UpdateOrder(ctx, order)
	} else if resp.StatusCode() == http.StatusTooManyRequests {
		return fmt.Errorf("number of requests to the service has been exceeded")
	} else if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("send order for calculation error: %s", string(resp.Body()))
	}
	var respBody struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float32 `json:"accrual"`
	}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return fmt.Errorf("unmarshal response body error: %s", string(resp.Body()))
	}
	if respBody.Status == "PROCESSED" {
		order.Status = models.OrderStatusProcessed
		order.Accrual = respBody.Accrual
		user, err := os.orderStore.SelectUserByID(ctx, order.UserID)
		if err != nil {
			return fmt.Errorf("fetching user error: %s", err.Error())
		}
		user.Balance += order.Accrual
		err = os.orderStore.UpdateUser(ctx, user)
		if err != nil {
			return fmt.Errorf("updating user error: %s", err.Error())
		}
	} else if respBody.Status == "INVALID" {
		order.Status = models.OrderStatusInvalid
	} else {
		order.Status = models.OrderStatusProcessing
	}
	return os.orderStore.UpdateOrder(ctx, order)
}
