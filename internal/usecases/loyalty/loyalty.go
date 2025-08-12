package loyalty

import (
	"context"

	"github.com/kdv2001/loyalty/internal/domain"
)

type loyaltyClient interface {
	GetAccruals(ctx context.Context, orderID domain.ID) error
}

type Implementation struct {
	loyaltyClient loyaltyClient
}

func NewImplementation(loyaltyClient loyaltyClient) *Implementation {
	return &Implementation{
		loyaltyClient: loyaltyClient,
	}
}
