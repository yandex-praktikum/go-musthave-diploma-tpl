package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Client struct {
	baseURL *url.URL
	client  *http.Client
}

type OrderStatus string

const (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type OrderAccrual struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual *float64    `json:"accrual,omitempty"`
}

type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %s", e.RetryAfter)
}

func New(rawURL string) (*Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse accrual url: %w", err)
	}
	return &Client{
		baseURL: u,
		client:  &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (c *Client) GetOrderInfo(ctx context.Context, number string) (*OrderAccrual, error) {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, "/api/orders", number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var oa OrderAccrual
		if err := json.NewDecoder(resp.Body).Decode(&oa); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &oa, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		ra := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{RetryAfter: ra}
	default:
		return nil, fmt.Errorf("unexpected status code %d from accrual system", resp.StatusCode)
	}
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return time.Minute
	}
	sec, err := strconv.Atoi(v)
	if err != nil || sec <= 0 {
		return time.Minute
	}
	return time.Duration(sec) * time.Second
}
