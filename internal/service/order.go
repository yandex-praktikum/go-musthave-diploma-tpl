package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/models"
	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/repository"
)

// Статусы заказа.
const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusRegistered = "REGISTERED" // из ответа системы начислений
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

// OrderStatusesPendingAccrual — статусы заказов, которые нужно опрашивать (только NEW — по ним ещё не получен финальный ответ).
var OrderStatusesPendingAccrual = []string{OrderStatusNew}

// OrderStatusesFromAccrual — допустимые статусы в ответе системы начислений (для валидации).
var OrderStatusesFromAccrual = map[string]bool{
	OrderStatusRegistered: true,
	OrderStatusInvalid:    true,
	OrderStatusProcessing: true,
	OrderStatusProcessed:  true,
}

// OrderStatusesFinalFromAccrual — финальные статусы: обновляем БД только при них (PROCESSED, INVALID).
var OrderStatusesFinalFromAccrual = map[string]bool{
	OrderStatusInvalid:   true,
	OrderStatusProcessed: true,
}

// OrderService — бизнес-логика заказов.
type OrderService struct {
	repo repository.OrderRepository
}

// NewOrderService создаёт сервис заказов.
func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// AddOrder добавляет номер заказа для пользователя.
func (s *OrderService) AddOrder(ctx context.Context, userID int64, number string) (*models.Order, bool, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil, false, &ErrValidation{Msg: "order number required"}
	}
	for _, r := range number {
		if !unicode.IsDigit(r) {
			return nil, false, &ErrValidation{Msg: "order number must contain only digits"}
		}
	}
	if !luhnValid(number) {
		return nil, false, &ErrValidation{Msg: "invalid order number format (Luhn)"}
	}

	order, err := s.repo.GetByNumber(ctx, number)
	if err != nil {
		var notFound *repository.ErrOrderNotFound
		if errors.As(err, &notFound) {
			newOrder, createErr := s.repo.Create(ctx, userID, number, OrderStatusNew)
			if createErr != nil {
				return nil, false, createErr
			}
			return newOrder, true, nil
		}
		return nil, false, err
	}

	if order.UserID == userID {
		return order, false, nil
	}
	return nil, false, &ErrOrderOwnedByOther{Number: number}
}

// ListOrders возвращает заказы пользователя по uploaded_at DESC.
func (s *OrderService) ListOrders(ctx context.Context, userID int64) ([]*models.Order, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// GetOrderNumbersPendingAccrual возвращает номера заказов в статусах, ожидающих опроса во внешней системе начислений.
func (s *OrderService) GetOrderNumbersPendingAccrual(ctx context.Context) ([]string, error) {
	return s.repo.ListNumbersPendingAccrual(ctx, OrderStatusesPendingAccrual)
}

// ApplyAccrualResult применяет результат из системы начислений к заказу только при финальном статусе (PROCESSED, INVALID).
// При REGISTERED или PROCESSING не обновляет БД (no-op, return nil).
func (s *OrderService) ApplyAccrualResult(ctx context.Context, number, status string, accrual *int) error {
	number = strings.TrimSpace(number)
	if number == "" {
		return &ErrValidation{Msg: "order number required"}
	}
	if !OrderStatusesFromAccrual[status] {
		return &ErrValidation{Msg: "invalid status from accrual: " + status}
	}
	if !OrderStatusesFinalFromAccrual[status] {
		return nil // REGISTERED, PROCESSING — не обновляем, опросим снова позже
	}
	if accrual != nil && *accrual < 0 {
		return &ErrValidation{Msg: "accrual must be non-negative"}
	}
	return s.repo.UpdateAccrualAndStatus(ctx, number, status, accrual)
}

// luhnValid проверяет номер по алгоритму Луна.
func luhnValid(number string) bool {
	var sum int
	parity := len(number) % 2
	for i, r := range number {
		if !unicode.IsDigit(r) {
			return false
		}
		d := int(r - '0')
		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return len(number) > 0 && sum%10 == 0
}
