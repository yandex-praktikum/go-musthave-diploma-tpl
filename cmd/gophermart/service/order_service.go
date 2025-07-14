package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type OrderRepo interface {
	CreateOrder(orderNumber string, userID int64) error
	GetOrderByNumber(orderNumber string) (*models.Order, error)
	GetOrderByNumberAndUserID(orderNumber string, userID int64) (*models.Order, error)
	GetOrdersByUserID(userID int64) ([]models.Order, error)
	GetOrdersForStatusUpdate() ([]models.Order, error)
	UpdateOrderStatus(orderID int64, status string) error
	AddBalanceTransaction(userID int64, orderID *int64, amount float64, txType string) error
	GetOrderAccrual(orderID int64) (*float64, error)
	GetUserBalance(userID int64) (current float64, withdrawn float64, err error)
	GetUserWithdrawals(userID int64) ([]models.WithdrawalResponse, error)
}

type OrderService struct {
	OrderRepo OrderRepo
	UserRepo  UserRepo
}

func NewOrderService(orderRepo OrderRepo, userRepo UserRepo) *OrderService {
	return &OrderService{OrderRepo: orderRepo, UserRepo: userRepo}
}

var (
	ErrOrderAlreadyUploadedByUser    = errors.New("order already uploaded by this user")
	ErrOrderAlreadyUploadedByAnother = errors.New("order already uploaded by another user")
	ErrInvalidOrderFormat            = errors.New("invalid order format")
	ErrInvalidOrderNumber            = errors.New("invalid order number")
	ErrInsufficientFunds             = errors.New("недостаточно средств")
)

func isValidLuhn(number string) bool {
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if double {
			digit = digit * 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

func (s *OrderService) UploadOrder(orderNumber string, userID int64) error {
	orderNumber = strings.TrimSpace(orderNumber)
	if orderNumber == "" {
		return ErrInvalidOrderFormat
	}
	for _, c := range orderNumber {
		if c < '0' || c > '9' {
			return ErrInvalidOrderNumber
		}
	}
	if !isValidLuhn(orderNumber) {
		return ErrInvalidOrderNumber
	}
	order, err := s.OrderRepo.GetOrderByNumber(orderNumber)
	if err == nil && order != nil {
		if order.UserID == userID {
			return ErrOrderAlreadyUploadedByUser
		} else {
			return ErrOrderAlreadyUploadedByAnother
		}
	}
	err = s.OrderRepo.CreateOrder(orderNumber, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetOrdersByUserID(userID int64) ([]models.Order, error) {
	return s.OrderRepo.GetOrdersByUserID(userID)
}

func (s *OrderService) StartOrderStatusWorker(ctx context.Context, accrualAddr string, logger *zap.Logger) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		client := &http.Client{Timeout: 5 * time.Second}
		for {
			select {
			case <-ctx.Done():
				logger.Info("Order status worker stopped")
				return
			case <-ticker.C:
				orders, err := s.OrderRepo.GetOrdersForStatusUpdate()
				if err != nil {
					logger.Error("Ошибка получения заказов для обновления статуса", zap.Error(err))
					continue
				}
				for _, order := range orders {
					url := fmt.Sprintf("%s/api/orders/%s", accrualAddr, order.OrderNumber)
					resp, err := client.Get(url)
					if err != nil {
						logger.Error("Ошибка запроса к accrual-сервису", zap.Error(err))
						continue
					}
					if resp.StatusCode == http.StatusNoContent {
						_ = s.OrderRepo.UpdateOrderStatus(order.ID, "INVALID")
						resp.Body.Close()
						continue
					}
					if resp.StatusCode != http.StatusOK {
						logger.Error("Неожиданный статус accrual-сервиса", zap.String("status", resp.Status))
						resp.Body.Close()
						continue
					}
					var accrualResp struct {
						Order   string   `json:"order"`
						Status  string   `json:"status"`
						Accrual *float64 `json:"accrual,omitempty"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
						logger.Error("Ошибка декодирования ответа accrual", zap.Error(err))
						resp.Body.Close()
						continue
					}
					resp.Body.Close()
					_ = s.OrderRepo.UpdateOrderStatus(order.ID, accrualResp.Status)
					if accrualResp.Accrual != nil && accrualResp.Status == "PROCESSED" {
						_ = s.OrderRepo.AddBalanceTransaction(order.UserID, &order.ID, *accrualResp.Accrual, "ACCRUAL")
					}
				}
			}
		}
	}()
}

func (s *OrderService) GetOrderAccrual(orderID int64) (*float64, error) {
	return s.OrderRepo.GetOrderAccrual(orderID)
}

func (s *OrderService) GetUserBalance(userID int64) (current float64, withdrawn float64, err error) {
	return s.OrderRepo.GetUserBalance(userID)
}

func (s *OrderService) WithdrawBalance(userID int64, orderNumber string, sum float64) error {
	if !isValidLuhn(orderNumber) {
		return ErrInvalidOrderNumber
	}
	current, _, err := s.GetUserBalance(userID)
	if err != nil {
		return err
	}
	if sum > current {
		return ErrInsufficientFunds
	}
	order, err := s.OrderRepo.GetOrderByNumber(orderNumber)
	if err != nil || order == nil {
		err = s.OrderRepo.CreateOrder(orderNumber, userID)
		if err != nil {
			return err
		}
		order, err = s.OrderRepo.GetOrderByNumber(orderNumber)
		if err != nil {
			return err
		}
	}
	return s.OrderRepo.AddBalanceTransaction(userID, &order.ID, sum, "WITHDRAWAL")
}

func (s *OrderService) GetUserWithdrawals(userID int64) ([]models.WithdrawalResponse, error) {
	return s.OrderRepo.GetUserWithdrawals(userID)
}
