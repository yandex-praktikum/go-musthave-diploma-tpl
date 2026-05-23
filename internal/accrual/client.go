// Package accrual предоставляет клиент для системы расчёта начислений.
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found in accrual system")

	ErrAccrualServerError = errors.New("accrual system server error")
)

// OrderStatus статус заказа в системе accrual
type OrderStatus string

const (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusProcessed  OrderStatus = "PROCESSED"
)

// OrderInfo информация о заказе из системы accrual
type OrderInfo struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual *float64    `json:"accrual,omitempty"`
}

// RateLimitError ошибка превышения лимита запросов
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %v", e.RetryAfter)
}

// Client клиент для системы расчёта начислений
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option функциональная опция для Client
type Option func(*Client)

// WithTimeout устанавливает таймаут для HTTP клиента
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient.Timeout = timeout
		}
	}
}

// NewClient создаёт новый клиент для системы accrual с опциями
func NewClient(baseURL string, opts ...Option) *Client {
	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30, // значение по умолчанию
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// GetOrderInfo получает информацию о заказе из системы accrual
func (c *Client) GetOrderInfo(ctx context.Context, number string) (*OrderInfo, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return c.parseOrderInfo(resp.Body)

	case http.StatusNoContent:
		return nil, ErrOrderNotFound

	case http.StatusTooManyRequests:
		retryAfter := c.parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{RetryAfter: retryAfter}

	case http.StatusInternalServerError:
		return nil, ErrAccrualServerError

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// parseOrderInfo парсит ответ системы accrual
func (c *Client) parseOrderInfo(body io.Reader) (*OrderInfo, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var info OrderInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &info, nil
}

// parseRetryAfter парсит заголовок Retry-After
func (c *Client) parseRetryAfter(headerValue string) time.Duration {
	if headerValue == "" {
		return time.Second * 60 // Значение по умолчанию
	}

	// Пробуем парсить как число секунд
	if seconds, err := strconv.Atoi(headerValue); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Пробуем парсить как HTTP дату
	if t, err := http.ParseTime(headerValue); err == nil {
		return time.Until(t)
	}

	return time.Second * 60
}
