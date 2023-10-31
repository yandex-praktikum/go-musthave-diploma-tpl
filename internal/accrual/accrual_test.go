package accrual

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	repo_mocks "github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository/mocks"
)

func TestServiceAccrual_ProcessedAccrualData(t *testing.T) {

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "valid",
			err:  nil,
		},
	}

	orders := []models.OrderResponse{
		models.OrderResponse{
			Number:  "371449635398431",
			Status:  "NEW",
			Accrual: 100,
		},
	}

	cfglog := &logger.Config{
		LogLevel: "info",
		DevMode:  true,
		Type:     "plaintext",
	}

	log := logger.NewAppLogger(cfglog)
	log.InitLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repo_mocks.NewMockOrders(ctrl)

			repo.EXPECT().GetOrdersWithStatus().Return(orders, nil)

			accrualService := NewServiceAccrual(repo, log, "http://127.0.0.1:8090")

			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()

			accrualService.ProcessedAccrualData(ctx)

			require.NoError(t, nil)
		})
	}
}
