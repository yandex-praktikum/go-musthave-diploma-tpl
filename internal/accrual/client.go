package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// accrualClient talks to the loyalty accrual system over HTTP.
type accrualClient struct {
	address string
	client  *http.Client
	logger  *zap.Logger
}

// NewAccrualClient creates an HTTP client for the loyalty accrual system at the given address.
func NewAccrualClient(address string, logger *zap.Logger) *accrualClient {
	return &accrualClient{
		address: address,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// GetStatus fetches the current accrual status of the order with the given number.
func (ac *accrualClient) GetStatus(ctx context.Context, orderNumber string) (accrualOrder, error) {
	var result accrualOrder
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/orders/%s", ac.address, orderNumber), nil)

	if err != nil {
		ac.logger.Error("error wrapping HTTP request", zap.Error(err))
		return result, err
	}

	response, err := ac.client.Do(request)
	if err != nil {
		ac.logger.Error("error sending HTTP request", zap.Error(err))
		return result, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			ac.logger.Error("error reading response body", zap.Error(err))
			return result, err
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			ac.logger.Error("error unmarshaling response body", zap.Error(err))
			return result, err
		}
		return result, nil
	}
	if response.StatusCode == http.StatusNoContent {
		ac.logger.Debug("order is not registered in the accrual system", zap.String("orderNumber", orderNumber))
		return result, ErrOrderNotRegistered
	}
	if response.StatusCode == http.StatusTooManyRequests {
		// TODO: honor the Retry-After header on 429 responses.
		ac.logger.Debug("accrual rate limit exceeded", zap.String("orderNumber", orderNumber))
		return result, ErrTooManyRequests
	}
	if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError {
		ac.logger.Error(ErrAccrualClient.Error(), zap.Int("status_code", response.StatusCode))
		return result, ErrAccrualClient
	}
	if response.StatusCode >= http.StatusInternalServerError {
		ac.logger.Error(ErrAccrualServer.Error(), zap.Int("status_code", response.StatusCode))
		return result, ErrAccrualServer
	}

	return result, ErrAccrualUnexpectedResponse
}
