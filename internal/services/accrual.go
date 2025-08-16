package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

// AccrualService сервис для работы с системой начисления баллов
type AccrualService struct {
	client  *retryablehttp.Client
	baseURL string
}

// NewAccrualService создает новый сервис начисления баллов
func NewAccrualService(baseURL string) *AccrualService {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 100 * time.Millisecond
	client.RetryWaitMax = 5 * time.Second
	client.HTTPClient.Timeout = 10 * time.Second

	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// Не делаем retry для rate limit
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			return false, nil
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	return &AccrualService{
		client:  client,
		baseURL: baseURL,
	}
}

// NewAccrualServiceWithRetry создает сервис с настраиваемыми параметрами retry
func NewAccrualServiceWithRetry(baseURL string, maxRetries int, baseDelay, maxDelay time.Duration) *AccrualService {
	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = baseDelay
	client.RetryWaitMax = maxDelay
	client.HTTPClient.Timeout = 10 * time.Second

	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			return false, nil
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	return &AccrualService{
		client:  client,
		baseURL: baseURL,
	}
}

// GetOrderInfo получает информацию о заказе из системы начисления
func (s *AccrualService) GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.baseURL, orderNumber)
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Выполняем запрос с retry
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp models.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &accrualResp, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		retryAfterStr := resp.Header.Get("Retry-After")
		var retryAfter time.Duration
		if retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}
	case http.StatusInternalServerError:
		return nil, ErrInternalServer
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
