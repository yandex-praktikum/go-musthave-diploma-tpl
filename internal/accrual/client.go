// Package accrual talks to the external loyalty accrual service.
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/true18/go-musthave-diploma-tpl/internal/money"
)

// External accrual statuses.
const (
	StatusRegistered = "REGISTERED"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

// ErrRateLimited is returned when accrual service responds with 429.
var ErrRateLimited = errors.New("accrual rate limited")

// ErrNoAddress is returned when accrual address is empty.
var ErrNoAddress = errors.New("accrual system address is empty")

// Result is the parsed response from the accrual service.
type Result struct {
	Order      string
	Status     string
	Accrual    *money.Amount
	NoContent  bool
	RetryAfter time.Duration
}

// Client fetches order accrual state over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// Fetch returns accrual information for one order.
func (c *Client) Fetch(ctx context.Context, number string) (Result, error) {
	if c.baseURL == "" {
		return Result{}, ErrNoAddress
	}

	endpoint := c.baseURL + "/api/orders/" + url.PathEscape(number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Result{}, fmt.Errorf("create accrual request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("fetch accrual: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result Result
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return Result{}, fmt.Errorf("decode accrual response: %w", err)
		}
		return result, nil
	case http.StatusNoContent:
		return Result{NoContent: true}, nil
	case http.StatusTooManyRequests:
		return Result{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}, ErrRateLimited
	default:
		return Result{}, fmt.Errorf("unexpected accrual status: %d", resp.StatusCode)
	}
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Second
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return time.Second
		}
		return time.Duration(seconds) * time.Second
	}

	when, err := http.ParseTime(value)
	if err != nil {
		return time.Second
	}
	duration := time.Until(when)
	if duration <= 0 {
		return time.Second
	}
	return duration
}
