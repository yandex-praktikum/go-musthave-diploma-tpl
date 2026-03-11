package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultRequestTimeout = 10 * time.Second

// Client — клиент к системе расчёта начислений.
type Client interface {
	GetOrder(ctx context.Context, orderNumber string) (*OrderResponse, error)
}

// HTTPClient — реализация Client через HTTP.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient создаёт HTTP-клиент к системе начислений. baseURL без завершающего слэша (например http://localhost:8080).
func NewHTTPClient(baseURL string, httpClient *http.Client) *HTTPClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultRequestTimeout}
	}
	return &HTTPClient{baseURL: baseURL, client: httpClient}
}

// GetOrder запрашивает информацию по заказу. 200 → *OrderResponse; 204 → *ErrOrderNotRegistered; 429 → *ErrRateLimit; 500/сеть → *ErrServerErrorContext (errors.Is(..., ErrServerError)).
func (c *HTTPClient) GetOrder(ctx context.Context, orderNumber string) (*OrderResponse, error) {
	url := c.baseURL + "/api/orders/" + orderNumber
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &ErrServerErrorContext{OrderNumber: orderNumber, URL: url, Err: fmt.Errorf("%w: %v", ErrServerError, err)}
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &ErrServerErrorContext{OrderNumber: orderNumber, URL: url, Err: fmt.Errorf("%w: %v", ErrServerError, err)}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var out OrderResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, &ErrServerErrorContext{OrderNumber: orderNumber, URL: url, Err: fmt.Errorf("%w: decode: %v", ErrServerError, err)}
		}
		return &out, nil
	case http.StatusNoContent:
		return nil, &ErrOrderNotRegistered{OrderNumber: orderNumber, URL: url}
	case http.StatusTooManyRequests:
		retryAfter := 60
		if s := resp.Header.Get("Retry-After"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				retryAfter = n
			}
		}
		return nil, &ErrRateLimit{OrderNumber: orderNumber, URL: url, RetryAfter: retryAfter}
	default:
		return nil, &ErrServerErrorContext{OrderNumber: orderNumber, URL: url, Err: ErrServerError}
	}
}
