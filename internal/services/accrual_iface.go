package services

import (
	"context"

	"github.com/vglushak/go-musthave-diploma-tpl/internal/models"
)

type AccrualServiceIface interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualResponse, error)
}
