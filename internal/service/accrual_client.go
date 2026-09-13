package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/b602op/go-musthave-diploma-tpl/internal/domain"
)

const (
	// maxRetries — максимальное число повторных попыток при 5xx.
	maxRetries = 3

	// baseRetryDelay — базовая задержка для экспоненциального backoff.
	baseRetryDelay = 500 * time.Millisecond

	// maxRetryDelay — верхняя граница задержки между попытками.
	maxRetryDelay = 5 * time.Second

	// defaultRetryAfter — значение по умолчанию для Retry-After при 429.
	defaultRetryAfter = 60 * time.Second
)

// AccrualClient — HTTP-клиент внешней системы расчёта баллов (accrual).
type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAccrualClient создаёт новый AccrualClient с базовым URL accrual-системы
// и таймаутом запросов 30 секунд.
func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetOrderInfo получает информацию о заказе от accrual-сервиса.
//
// При ответах 5xx выполняет до maxRetries повторных попыток
// с экспоненциальной задержкой. При ответе 429 возвращает
// *domain.RateLimitError со временем паузы из заголовка Retry-After.
func (c *AccrualClient) GetOrderInfo(orderNumber string) (*domain.AccrualResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(attempt)
			log.Printf("[DEBUG] accrual retry %d/%d for order %s after %v",
				attempt, maxRetries, orderNumber, delay)
			time.Sleep(delay)
		}

		resp, err := c.doRequest(orderNumber)
		if err != nil {
			// Сетевые ошибки (таймаут, connection refused) — ретраим
			lastErr = err
			continue
		}

		// 5xx — временный сбой сервера, ретраим
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("accrual returned %d", resp.StatusCode)
			continue
		}

		// 4xx (кроме 429) — постоянная ошибка, не ретраим
		if resp.StatusCode >= 400 && resp.StatusCode != http.StatusTooManyRequests {
			return nil, fmt.Errorf("unexpected status code from accrual: %d", resp.StatusCode)
		}

		// 2xx/204/429 — обрабатываем и выходим
		return c.parseResponse(orderNumber, resp)
	}

	return nil, fmt.Errorf("accrual request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// doRequest выполняет один HTTP-запрос к accrual.
func (c *AccrualClient) doRequest(orderNumber string) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call accrual service: %w", err)
	}

	return resp, nil
}

// parseResponse разбирает ответ accrual-сервиса.
// Закрывает тело ответа.
func (c *AccrualClient) parseResponse(orderNumber string, resp *http.Response) (*domain.AccrualResponse, error) {
	defer resp.Body.Close()

	// 204 No Content — заказ не зарегистрирован
	if resp.StatusCode == http.StatusNoContent {
		return &domain.AccrualResponse{
			Order:  orderNumber,
			Status: domain.AccrualStatusRegistered,
		}, nil
	}

	// 200 OK — заказ найден
	if resp.StatusCode == http.StatusOK {
		var result domain.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode accrual response: %w", err)
		}
		return &result, nil
	}

	// 429 Too Many Requests — парсим Retry-After
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &domain.RateLimitError{RetryAfter: retryAfter}
	}

	return nil, fmt.Errorf("unexpected status code from accrual: %d", resp.StatusCode)
}

// backoffDelay рассчитывает задержку для попытки attempt (attempt >= 1)
// по формуле min(baseRetryDelay * 2^(attempt-1), maxRetryDelay).
func backoffDelay(attempt int) time.Duration {
	delay := time.Duration(float64(baseRetryDelay) * math.Pow(2, float64(attempt-1)))
	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}
	return delay
}

// parseRetryAfter разбирает заголовок Retry-After в time.Duration.
//
// Поддерживает два формата (RFC 7231):
//   - целое число секунд ("60");
//   - HTTP-дата ("Wed, 21 Oct 2015 07:28:00 GMT").
//
// Если заголовок пуст, имеет неверный формат или указывает время в прошлом,
// возвращает 60 секунд по умолчанию и логирует причину.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		log.Println("[DEBUG] parseRetryAfter: empty header, using default 60s")
		return defaultRetryAfter
	}

	// Формат 1: число секунд
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			log.Printf("[WARN] parseRetryAfter: non-positive value %q, using default 60s", value)
			return defaultRetryAfter
		}
		return time.Duration(seconds) * time.Second
	}

	// Формат 2: HTTP-дата
	if t, err := http.ParseTime(value); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
		log.Printf("[WARN] parseRetryAfter: date %q is in the past, using default 60s", value)
		return defaultRetryAfter
	}

	log.Printf("[WARN] parseRetryAfter: cannot parse %q, using default 60s", value)
	return defaultRetryAfter
}
