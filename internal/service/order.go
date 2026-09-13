package service

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
	"github.com/b602op/go-musthave-diploma-tpl/internal/repository"
	"github.com/b602op/go-musthave-diploma-tpl/internal/utils"
)

// Ограничиваем число одновременных запросов к accrual.
const maxConcurrency = 10

// OrderService — сервис загрузки и обработки заказов.
type OrderService struct {
	orderRepo     *repository.OrderRepository
	balanceRepo   *repository.BalanceRepository
	accrualClient *AccrualClient
}

// NewOrderService создаёт новый OrderService с указанными репозиториями
// заказов, балансов и клиентом accrual-системы.
func NewOrderService(
	orderRepo *repository.OrderRepository,
	balanceRepo *repository.BalanceRepository,
	accrualClient *AccrualClient,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
	}
}

// UploadOrder загружает номер заказа
func (s *OrderService) UploadOrder(userID int64, orderNumber string) error {
	// Проверяем номер заказа по алгоритму Луна
	if !utils.ValidLuhn(orderNumber) {
		return domain.ErrInvalidOrderNumber
	}

	// Проверяем, существует ли заказ
	existing, err := s.orderRepo.FindByNumber(orderNumber)
	if err != nil && !errors.Is(err, domain.ErrOrderNotFound) {
		return fmt.Errorf("failed to check order existence: %w", err)
	}

	if existing != nil {
		if existing.UserID == userID {
			return domain.ErrOrderAlreadyUploadedByUser
		}
		return domain.ErrOrderAlreadyUploadedByAnotherUser
	}

	// Создаем новый заказ
	order := &domain.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     domain.OrderStatusNew,
		UploadedAt: time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.orderRepo.Create(order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetUserOrders возвращает все заказы пользователя
func (s *OrderService) GetUserOrders(userID int64) ([]*domain.Order, error) {
	orders, err := s.orderRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	return orders, nil
}

// ProcessPendingOrders обрабатывает заказы, ожидающие расчета
// ProcessPendingOrders обрабатывает заказы, ожидающие расчета.
//
// Заказы обрабатываются параллельно с ограничением количества
// одновременных запросов к accrual через буферизованный канал-семафор.
// Ошибки отдельных заказов логируются и не прерывают обработку остальных.
func (s *OrderService) ProcessPendingOrders(limit int) error {
	orders, err := s.orderRepo.GetPendingOrders(limit)
	if err != nil {
		return fmt.Errorf("failed to get pending orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		sem <- struct{}{} // занимаем слот (блокируется, если все заняты)

		go func(order *domain.Order) {
			defer wg.Done()
			defer func() { <-sem }() // освобождаем слот

			if err := s.processOrder(order); err != nil {
				log.Printf("[ERROR] processOrder: order=%s, err=%v", order.Number, err)
			}
		}(order)
	}

	wg.Wait()
	return nil
}

// processOrder обрабатывает один заказ.
//
// Получает информацию о заказе из accrual-сервиса, обновляет статус и,
// если начисление положительное, добавляет баллы на баланс пользователя.
// Возвращает ошибку с контекстом — логирование выполняет вызывающий код.
func (s *OrderService) processOrder(order *domain.Order) error {
	// Получаем информацию от accrual
	resp, err := s.accrualClient.GetOrderInfo(order.Number)
	if err != nil {
		return fmt.Errorf("get order info from accrual for %s: %w", order.Number, err)
	}

	// Маппим статусы
	var newStatus string
	var accrual *float64

	switch resp.Status {
	case domain.AccrualStatusRegistered:
		newStatus = domain.OrderStatusProcessing
	case domain.AccrualStatusProcessing:
		newStatus = domain.OrderStatusProcessing
	case domain.AccrualStatusProcessed:
		newStatus = domain.OrderStatusProcessed
		if resp.Accrual != nil && *resp.Accrual > 0 {
			accrual = resp.Accrual
		}
	case domain.AccrualStatusInvalid:
		newStatus = domain.OrderStatusInvalid
	default:
		newStatus = domain.OrderStatusProcessing
	}

	// Обновляем заказ
	order.Status = newStatus
	order.Accrual = accrual
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(order); err != nil {
		return fmt.Errorf("update order %s: %w", order.Number, err)
	}

	// Если заказ обработан и есть начисление — добавляем баллы
	if newStatus == domain.OrderStatusProcessed && accrual != nil && *accrual > 0 {
		if err := s.balanceRepo.AddBalance(order.UserID, *accrual); err != nil {
			return fmt.Errorf("add balance for order %s: %w", order.Number, err)
		}
	}

	return nil
}
