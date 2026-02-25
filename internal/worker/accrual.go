// Package worker реализует фоновые воркеры
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/anon-d/gophermarket/internal/repository"
)

// WorkerRepository интерфейс репозитория, используемый воркером
//
//go:generate go run go.uber.org/mock/mockgen -destination=mock_repository_test.go -package=worker github.com/anon-d/gophermarket/internal/worker WorkerRepository
type WorkerRepository interface {
	GetOrdersForProcessing(ctx context.Context) ([]repository.Order, error)
	StreamOrdersForProcessing(ctx context.Context) iter.Seq[repository.Order]
	UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual float64) error
}

// AccrualResponse ответ от системы расчёта начислений
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// AccrualWorker воркер для проверки начислений
type AccrualWorker struct {
	repo           WorkerRepository
	accrualAddress string
	logger         *zap.Logger
	httpClient     *http.Client
	interval       time.Duration
}

// NewAccrualWorker создаёт новый воркер
func NewAccrualWorker(repo WorkerRepository, accrualAddress string, logger *zap.Logger) *AccrualWorker {
	return &AccrualWorker{
		repo:           repo,
		accrualAddress: accrualAddress,
		logger:         logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		interval: 5 * time.Second,
	}
}

// Start запускает воркер
func (w *AccrualWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Accrual worker остановлен")
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

// processOrders обрабатывает заказы, ожидающие проверки
func (w *AccrualWorker) processOrders(ctx context.Context) {
	// Используем iter.Seq для стриминга заказов
	for order := range w.repo.StreamOrdersForProcessing(ctx) {
		select {
		case <-ctx.Done():
			return
		default:
			w.processOrder(ctx, order.Number)
		}
	}
}

// processOrder обрабатывает один заказ
func (w *AccrualWorker) processOrder(ctx context.Context, orderNumber string) {
	url := fmt.Sprintf("%s/api/orders/%s", w.accrualAddress, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		w.logger.Error("Ошибка создания запроса", zap.String("order", orderNumber), zap.Error(err))
		return
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.logger.Error("Ошибка запроса к accrual сервису", zap.String("order", orderNumber), zap.Error(err))
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			w.logger.Error("Ошибка декодирования ответа", zap.String("order", orderNumber), zap.Error(err))
			return
		}

		// Обновляем статус заказа
		if err := w.repo.UpdateOrderStatus(ctx, orderNumber, accrualResp.Status, accrualResp.Accrual); err != nil {
			w.logger.Error("Ошибка обновления статуса заказа", zap.String("order", orderNumber), zap.Error(err))
			return
		}

		w.logger.Info("Заказ обработан",
			zap.String("order", orderNumber),
			zap.String("status", accrualResp.Status),
			zap.Float64("accrual", accrualResp.Accrual))

	case http.StatusNoContent:
		// Заказ не зарегистрирован в системе расчёта - ничего не делаем
		w.logger.Debug("Заказ не найден в accrual сервисе", zap.String("order", orderNumber))

	case http.StatusTooManyRequests:
		// Превышен лимит запросов - ждём
		retryAfter := resp.Header.Get("Retry-After")
		w.logger.Warn("Превышен лимит запросов к accrual сервису",
			zap.String("order", orderNumber),
			zap.String("retry_after", retryAfter))

		if retryAfter != "" {
			if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
				time.Sleep(seconds)
			}
		} else {
			time.Sleep(60 * time.Second)
		}

	default:
		w.logger.Error("Неожиданный код ответа от accrual сервиса",
			zap.String("order", orderNumber),
			zap.Int("status_code", resp.StatusCode))
	}
}
