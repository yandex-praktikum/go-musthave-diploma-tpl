package orderclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client представляет клиент API заказов
type Client struct {
	config     Config
	httpClient *http.Client
}

// OrderClient Интерфейс клиента для упрощения тестирования
type OrderClient interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*OrderResponse, error)
	GetOrderInfoWithRetry(ctx context.Context, orderNumber string, retryOpts ...RetryOption) (*OrderResponse, error)
}

// Option функция для настройки клиента
type Option func(*Client) error

// RetryOption параметры для повторных попыток
type RetryOption func(*retryOptions)

type retryOptions struct {
	maxRetries  int
	retryDelay  time.Duration
	maxWaitTime time.Duration
	retryOn     []int // статус коды для повторных попыток
}

// DefaultRetryOptions дефолтные настройки для повторных попыток
var DefaultRetryOptions = retryOptions{
	maxRetries:  3,
	retryDelay:  1 * time.Second,
	maxWaitTime: 30 * time.Second,
	retryOn:     []int{429, 500, 502, 503, 504},
}

// WithMaxRetries устанавливает максимальное количество повторных попыток
func WithMaxRetries(maxRetries int) RetryOption {
	return func(opts *retryOptions) {
		if maxRetries < 0 {
			maxRetries = 0
		}
		opts.maxRetries = maxRetries
	}
}

// WithRetryDelay устанавливает задержку между попытками
func WithRetryDelay(delay time.Duration) RetryOption {
	return func(opts *retryOptions) {
		if delay < 0 {
			delay = 0
		}
		opts.retryDelay = delay
	}
}

// WithMaxWaitTime устанавливает максимальное время ожидания всех попыток
func WithMaxWaitTime(maxWaitTime time.Duration) RetryOption {
	return func(opts *retryOptions) {
		if maxWaitTime <= 0 {
			maxWaitTime = 0
		}
		opts.maxWaitTime = maxWaitTime
	}
}

// WithRetryOn устанавливает статус коды для повторных попыток
func WithRetryOn(statusCodes ...int) RetryOption {
	return func(opts *retryOptions) {
		opts.retryOn = statusCodes
	}
}

// Validate проверяет конфигурацию на корректность
func (c Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("BaseURL cannot be empty")
	}

	if _, err := url.Parse(c.BaseURL); err != nil {
		return fmt.Errorf("invalid BaseURL: %w", err)
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("Timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries cannot be negative")
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("RetryDelay cannot be negative")
	}

	if c.UserAgent == "" {
		return fmt.Errorf("UserAgent cannot be empty")
	}

	if c.MaxResponseSize <= 0 {
		return fmt.Errorf("MaxResponseSize must be positive")
	}

	if c.MaxResponseSize > 100*1024*1024 { // 100 MB
		return fmt.Errorf("MaxResponseSize is too large")
	}

	return nil
}

// New создает новый экземпляр клиента
func New(config Config, opts ...Option) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

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
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return client, nil
}

// NewWithDefaults создает клиент с настройками по умолчанию
func NewWithDefaults(opts ...Option) (*Client, error) {
	return New(DefaultConfig(), opts...)
}

// WithHTTPClient устанавливает кастомный HTTP клиент
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("httpClient cannot be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithBaseURL устанавливает базовый URL
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		if baseURL == "" {
			return fmt.Errorf("baseURL cannot be empty")
		}
		if _, err := url.Parse(baseURL); err != nil {
			return fmt.Errorf("invalid baseURL: %w", err)
		}
		c.config.BaseURL = baseURL
		return nil
	}
}

// WithTimeout устанавливает таймаут
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout <= 0 {
			return fmt.Errorf("timeout must be positive")
		}
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithUserAgent устанавливает User-Agent
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		if userAgent == "" {
			return fmt.Errorf("userAgent cannot be empty")
		}
		c.config.UserAgent = userAgent
		return nil
	}
}

// WithMaxResponseSize устанавливает максимальный размер ответа
func WithMaxResponseSize(maxSize int) Option {
	return func(c *Client) error {
		if maxSize <= 0 {
			return fmt.Errorf("maxResponseSize must be positive")
		}
		if maxSize > 100*1024*1024 { // 100 MB
			return fmt.Errorf("maxResponseSize is too large")
		}
		c.config.MaxResponseSize = maxSize
		return nil
	}
}

// GetOrderInfo получает информацию о заказе
func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderResponse, error) {
	return c.getOrderInfoWithRetry(ctx, orderNumber, DefaultRetryOptions)
}

// GetOrderInfoWithRetry получает информацию о заказе с кастомными параметрами повторных попыток
func (c *Client) GetOrderInfoWithRetry(ctx context.Context, orderNumber string, retryOpts ...RetryOption) (*OrderResponse, error) {
	opts := DefaultRetryOptions
	for _, opt := range retryOpts {
		opt(&opts)
	}

	return c.getOrderInfoWithRetry(ctx, orderNumber, opts)
}

// getOrderInfoWithRetry внутренний метод с поддержкой повторных попыток
func (c *Client) getOrderInfoWithRetry(ctx context.Context, orderNumber string, opts retryOptions) (*OrderResponse, error) {
	var lastErr error

	// Создаем контекст с общим таймаутом для всех попыток
	if opts.maxWaitTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.maxWaitTime)
		defer cancel()
	}

	attempt := 0
	for {
		// Проверяем, не отменен ли контекст перед каждой попыткой
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, err
		}

		// Выполняем запрос
		response, err := c.executeRequest(ctx, orderNumber)

		// Если запрос успешен, возвращаем результат
		if err == nil {
			return response, nil
		}

		// Сохраняем последнюю ошибку
		lastErr = err

		// Проверяем, нужно ли делать повторную попытку
		shouldRetry, delay := c.shouldRetry(err, attempt, opts)
		if !shouldRetry {
			return nil, err
		}

		// Увеличиваем счетчик попыток
		attempt++

		// Ждем перед следующей попыткой
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				if lastErr != nil {
					return nil, lastErr
				}
				return nil, ctx.Err()
			case <-timer.C:
				// Продолжаем цикл
			}
		}
	}
}

// executeRequest выполняет одиночный HTTP запрос с ограничением размера ответа
func (c *Client) executeRequest(ctx context.Context, orderNumber string) (*OrderResponse, error) {
	// Экранируем номер заказа для безопасности URL
	escapedOrderNumber := url.PathEscape(orderNumber)

	// Формируем URL запроса
	sprintf := fmt.Sprintf("%s/api/orders/%s", c.config.BaseURL, escapedOrderNumber)

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sprintf, nil)
	if err != nil {
		return nil, &APIError{
			Type:        ErrorTypeNetwork,
			Message:     "failed to create request",
			OriginalErr: err,
		}
	}

	// Устанавливаем заголовки
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Проверяем, была ли ошибка из-за таймаута контекста
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, &APIError{
				Type:        ErrorTypeTimeout,
				Message:     "request timed out",
				OriginalErr: ctxErr,
			}
		}

		return nil, &APIError{
			Type:        ErrorTypeNetwork,
			Message:     "request failed",
			OriginalErr: err,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Читаем тело ответа с ограничением размера
	limitedReader := io.LimitReader(resp.Body, int64(c.config.MaxResponseSize))
	body, errReadAll := io.ReadAll(limitedReader)
	if errReadAll != nil {
		return nil, &APIError{
			Type:        ErrorTypeNetwork,
			StatusCode:  resp.StatusCode,
			Message:     "failed to read response body",
			OriginalErr: err,
		}
	}

	// Проверяем, не был ли ответ обрезан
	if len(body) == c.config.MaxResponseSize {
		// Пытаемся прочитать еще один байт, чтобы убедиться, что ответ был обрезан
		var extraByte [1]byte
		n, _ := resp.Body.Read(extraByte[:])
		if n > 0 {
			return nil, &APIError{
				Type:       ErrorTypeResponseTooLarge,
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("response exceeds maximum size of %d bytes", c.config.MaxResponseSize),
			}
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

// shouldRetry определяет, нужно ли делать повторную попытку
func (c *Client) shouldRetry(err error, attempt int, opts retryOptions) (bool, time.Duration) {
	// Если достигнуто максимальное количество попыток, не повторяем
	if attempt >= opts.maxRetries {
		return false, 0
	}

	var apiErr *APIError
	ok := errors.As(err, &apiErr)
	if !ok {
		// Если это не APIError, не повторяем
		return false, 0
	}

	// Не повторяем для ошибок таймаута
	if apiErr.Type == ErrorTypeTimeout {
		return false, 0
	}

	// Проверяем, нужно ли повторять для данного статус кода
	shouldRetryStatusCode := false
	for _, statusCode := range opts.retryOn {
		if apiErr.StatusCode == statusCode {
			shouldRetryStatusCode = true
			break
		}
	}

	if !shouldRetryStatusCode {
		return false, 0
	}

	// Вычисляем базовую задержку
	delay := opts.retryDelay

	// Для rate limit используем заголовок Retry-After
	if apiErr.Type == ErrorTypeRateLimit && apiErr.RetryAfter != "" {
		if retryDelay, parseErr := parseRetryAfter(apiErr.RetryAfter); parseErr == nil {
			delay = retryDelay
		}
	}

	// Экспоненциальная задержка
	delay = delay * (1 << uint(attempt)) // Умножаем на 2^attempt

	// Добавляем случайный джиттер (±20%)
	// rand.Float64() безопасен для конкурентного доступа в Go 1.20+
	jitterPercent := 0.4 // ±20%
	jitterMultiplier := 1 + (rand.Float64()*jitterPercent - jitterPercent/2)
	delay = time.Duration(float64(delay) * jitterMultiplier)

	// Ограничиваем задержку снизу (минимум 10 мс)
	if delay < 10*time.Millisecond {
		delay = 10 * time.Millisecond
	}

	// Ограничиваем задержку сверху
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}

	return true, delay
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

	// Пробуем парсить как секунды (целое число)
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		if seconds < 0 {
			return 0, fmt.Errorf("retry-after seconds cannot be negative")
		}
		if seconds > 86400 { // 24 часа
			return 0, fmt.Errorf("retry-after seconds too large: %d", seconds)
		}
		return time.Duration(seconds) * time.Second, nil
	}

	// Пробуем парсить как RFC1123 дату
	if date, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return calculateRetryDuration(date)
	}

	// Пробуем парсить как RFC1123Z (с часовым поясом)
	if date, err := time.Parse(time.RFC1123Z, retryAfter); err == nil {
		return calculateRetryDuration(date)
	}

	// Пробуем парсить как RFC3339
	if date, err := time.Parse(time.RFC3339, retryAfter); err == nil {
		return calculateRetryDuration(date)
	}

	return 0, fmt.Errorf("invalid retry-after format: %s", retryAfter)
}

// calculateRetryDuration вычисляет длительность до указанной даты
func calculateRetryDuration(date time.Time) (time.Duration, error) {
	now := time.Now()
	if date.After(now) {
		duration := date.Sub(now)
		// Ограничиваем максимальную задержку (например, 24 часа)
		if duration > 24*time.Hour {
			return 0, fmt.Errorf("retry-after date too far in the future: %v", date)
		}
		return duration, nil
	}
	return 0, nil
}

// Config Конфигурация с максимальным размером ответа
type Config struct {
	BaseURL         string
	Timeout         int // Секунды, храним только для конфигурации
	MaxRetries      int
	RetryDelay      int
	UserAgent       string
	MaxResponseSize int // в байтах
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		BaseURL:         "http://localhost:8081",
		Timeout:         30,
		MaxRetries:      3,
		RetryDelay:      1,
		UserAgent:       "OrderClient/1.0",
		MaxResponseSize: 10 * 1024 * 1024, // 10 MB
	}
}
