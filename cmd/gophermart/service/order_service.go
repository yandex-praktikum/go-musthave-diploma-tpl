package service

import (
	"errors"
	"strings"

	"github.com/AlexeySalamakhin/gophermart/cmd/gophermart/models"
)

type OrderRepo interface {
	CreateOrder(orderNumber string, userID int64) error
	GetOrderByNumber(orderNumber string) (*models.Order, error)
	GetOrderByNumberAndUserID(orderNumber string, userID int64) (*models.Order, error)
	GetOrdersByUserID(userID int64) ([]models.Order, error)
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
