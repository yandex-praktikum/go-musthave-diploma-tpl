package service

import (
	"context"
	"log"
	"time"

	"gophermart/internal/accrual"
	"gophermart/internal/repository"
)

type AccrualService struct {
	orderRepo     *repository.OrderRepository
	accrualClient *accrual.Client
}

func NewAccrualService(orderRepo *repository.OrderRepository, accrualClient *accrual.Client) *AccrualService {
	return &AccrualService{
		orderRepo:     orderRepo,
		accrualClient: accrualClient,
	}
}

func (s *AccrualService) StartWorker(ctx context.Context) {
	if s.accrualClient == nil {
		log.Println("Accrual worker: no accrual client, skipping")
		return
	}

	log.Println("Accrual worker started")
	defer log.Println("Accrual worker stopped")

	pollInterval := time.Second
	batchSize := 100

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollInterval):
		}

		workerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		orders, err := s.orderRepo.GetPendingOrders(workerCtx, batchSize)
		cancel()

		if err != nil {
			log.Printf("Accrual worker: query orders error: %v", err)
			continue
		}

		if len(orders) == 0 {
			pollInterval = min(pollInterval*2, 10*time.Second)
			continue
		}

		pollInterval = time.Second
		for _, ord := range orders {
			select {
			case <-ctx.Done():
				return
			default:
				s.processOrder(ctx, ord.ID, ord.Number, ord.UserID)
			}
		}
	}
}

func (s *AccrualService) processOrder(ctx context.Context, orderID int64, number string, userID int64) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := s.accrualClient.GetOrderInfo(ctx, number)
	if err != nil {
		if rl, ok := err.(*accrual.RateLimitError); ok {
			log.Printf("Accrual worker: rate limit, sleeping %s", rl.RetryAfter)
			time.Sleep(rl.RetryAfter)
			return
		}
		log.Printf("Accrual worker: get order info error: %v", err)
		return
	}

	if info == nil {
		return
	}

	switch info.Status {
	case accrual.StatusRegistered, accrual.StatusProcessing:
		if err := s.orderRepo.UpdateStatus(ctx, orderID, "PROCESSING"); err != nil {
			log.Printf("Accrual worker: update order PROCESSING error: %v", err)
		}
	case accrual.StatusInvalid:
		if err := s.orderRepo.UpdateStatus(ctx, orderID, "INVALID"); err != nil {
			log.Printf("Accrual worker: update order INVALID error: %v", err)
		}
	case accrual.StatusProcessed:
		var accrualVal float64
		if info.Accrual != nil {
			accrualVal = *info.Accrual
		}

		if err := s.orderRepo.UpdateStatusWithAccrual(ctx, orderID, "PROCESSED", accrualVal); err != nil {
			log.Printf("Accrual worker: update order PROCESSED error: %v", err)
		}
	}
}
