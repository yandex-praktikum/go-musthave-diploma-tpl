package accrual

import (
	"net/http"
	"time"
)

// accrualClient talks to the loyalty accrual system over HTTP.
type accrualClient struct {
	address string
	client  *http.Client
}

// NewAccrualClient creates an HTTP client for the loyalty accrual system at the given address.
func NewAccrualClient(address string) *accrualClient {
	return &accrualClient{
		address: address,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
