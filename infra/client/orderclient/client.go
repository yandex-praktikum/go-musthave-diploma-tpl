package orderclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client представляет клиент API заказов
type Client struct {
	config     Config
	httpClient *http.Client
	userAgent  string
}

// Option функция для настройки клиента
type Option func(*Client)

// New создает новый экземпляр клиента
func New(config Config, opts ...Option) *Client {
	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		userAgent: config.UserAgent,
	}

	// Применяем опции
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// NewWithDefaults создает клиент с настройками по умолчанию
func NewWithDefaults(opts ...Option) *Client {
	return New(DefaultConfig(), opts...)
}

// WithHTTPClient устанавливает кастомный HTTP клиент
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL устанавливает базовый URL
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.config.BaseURL = baseURL
	}
}

// WithTimeout устанавливает таймаут
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent устанавливает User-Agent
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// GetOrderInfo получает информацию о заказе
func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderResponse, error) {
	return c.getOrderInfoWithRetry(ctx, orderNumber, 0)
}

// getOrderInfoWithRetry внутренний метод с поддержкой повторных попыток
func (c *Client) getOrderInfoWithRetry(ctx context.Context, orderNumber string, attempt int) (*OrderResponse, error) {
	// Формируем URL запроса
	url := fmt.Sprintf("%s/api/orders/%s", c.config.BaseURL, orderNumber)

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &APIError{
			Type:        ErrorTypeNetwork,
			Message:     "failed to create request",
			OriginalErr: err,
		}
	}

	// Устанавливаем заголовки
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &APIError{
			Type:        ErrorTypeNetwork,
			Message:     "request failed",
			OriginalErr: err,
		}
	}
	defer func(Body io.ReadCloser) {
		errClose := Body.Close()
		if errClose != nil {
			return
		}
	}(resp.Body)

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			Type:        ErrorTypeNetwork,
			StatusCode:  resp.StatusCode,
			Message:     "failed to read response body",
			OriginalErr: err,
		}
	}

	// Обрабатываем статус код
	switch resp.StatusCode {
	case http.StatusOK:
		return c.parseOrderResponse(body)

	case http.StatusNoContent:
		return nil, &APIError{
			Type:       ErrorTypeNotFound,
			StatusCode: http.StatusNoContent,
			Message:    fmt.Sprintf("order %s not found", orderNumber),
		}

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")

		// Если есть попытки и можно повторить
		if attempt < c.config.MaxRetries {
			// Ждем перед повторной попыткой
			delay := time.Duration(c.config.RetryDelay) * time.Second
			if retryDelay, err := parseRetryAfter(retryAfter); err == nil {
				delay = retryDelay
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				return c.getOrderInfoWithRetry(ctx, orderNumber, attempt+1)
			}
		}

		return nil, &APIError{
			Type:       ErrorTypeRateLimit,
			StatusCode: http.StatusTooManyRequests,
			Message:    "rate limit exceeded",
			RetryAfter: retryAfter,
		}

	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return nil, &APIError{
			Type:       ErrorTypeClient,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("client error: %s", string(body)),
		}

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		// Для серверных ошибок пробуем повторить
		if attempt < c.config.MaxRetries {
			time.Sleep(time.Duration(c.config.RetryDelay) * time.Second)
			return c.getOrderInfoWithRetry(ctx, orderNumber, attempt+1)
		}

		return nil, &APIError{
			Type:       ErrorTypeServer,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("server error: %s", string(body)),
		}

	default:
		return nil, &APIError{
			Type:       ErrorTypeClient,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}
}

// parseOrderResponse парсит успешный ответ
func (c *Client) parseOrderResponse(body []byte) (*OrderResponse, error) {
	var response OrderResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, &APIError{
			Type:        ErrorTypeParse,
			Message:     "failed to parse JSON response",
			OriginalErr: err,
		}
	}

	// Валидируем статус заказа
	if !response.Status.IsValid() {
		return nil, &APIError{
			Type:    ErrorTypeParse,
			Message: fmt.Sprintf("invalid order status: %s", response.Status),
		}
	}

	return &response, nil
}

// parseRetryAfter парсит заголовок Retry-After
func parseRetryAfter(retryAfter string) (time.Duration, error) {
	if retryAfter == "" {
		return 0, fmt.Errorf("empty retry-after")
	}

	// Пробуем парсить как секунды
	if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
		return seconds, nil
	}

	// Пробуем парсить как RFC1123 дату
	if date, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return time.Until(date), nil
	}

	return 0, fmt.Errorf("invalid retry-after format: %s", retryAfter)
}

// GetOrderInfoWithRetry получает информацию о заказе с кастомным количеством попыток
func (c *Client) GetOrderInfoWithRetry(ctx context.Context, orderNumber string, maxRetries int) (*OrderResponse, error) {
	// Сохраняем оригинальное значение
	originalRetries := c.config.MaxRetries
	c.config.MaxRetries = maxRetries
	defer func() { c.config.MaxRetries = originalRetries }()

	return c.getOrderInfoWithRetry(ctx, orderNumber, 0)
}
