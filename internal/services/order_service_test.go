package services_test

import (
	"context"
	"errors"
	"testing"

	clientMocks "github.com/KalessinD/gophermart/internal/clients/mocks"
	"github.com/KalessinD/gophermart/internal/common"
	"github.com/KalessinD/gophermart/internal/middleware"
	model "github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/processors"
	repoMocks "github.com/KalessinD/gophermart/internal/repositories/postgresql/mocks"
	"github.com/KalessinD/gophermart/internal/services"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func processTaskChannel(t *testing.T, ch <-chan *processors.Task) {
	t.Helper()

	go func() {
		for {
			select {
			case _, opened := <-ch:
				if !opened {
					return
				}
			case <-t.Context().Done():
				return
			}
		}
	}()
}

func getCtxWithClaims(t *testing.T, userID string) context.Context {
	t.Helper()
	claims := &common.Claims{
		UserID: userID,
	}
	return context.WithValue(context.Background(), middleware.ClaimsKey, claims)
}

func TestOrderActions_Store(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	outCh := make(chan *processors.Task)

	processTaskChannel(t, outCh)

	mockRepo := repoMocks.NewMockSQLStorageInterface(ctrl)
	mockClient := clientMocks.NewMockAccrualClienttInterface(ctrl)
	service := services.NewOrderService(mockRepo, mockClient, outCh)

	validOrderID := "79927398713"
	userID := "user_1"
	ctx := getCtxWithClaims(t, userID)

	tests := []struct {
		name       string
		orderID    string
		mockExpect func()
		wantErr    error
	}{
		{
			name:    "Success - new order",
			orderID: validOrderID,
			mockExpect: func() {
				mockRepo.EXPECT().
					AddOrder(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "Error - invalid Luhn format",
			orderID: "123",
			mockExpect: func() {
			},
			wantErr: model.ErrOrderWrongFormat,
		},
		{
			name:    "Error - order belongs to other user",
			orderID: validOrderID,
			mockExpect: func() {
				mockRepo.EXPECT().
					AddOrder(gomock.Any(), gomock.Any()).
					Return(model.ErrOrderExists)

				mockRepo.EXPECT().
					GetOrder(gomock.Any(), validOrderID, userID).
					Return(nil, model.ErrOrderNotFound)
			},
			wantErr: model.ErrOrderBelongsToOtherUser,
		},
		{
			name:    "Error - order already exists for current user",
			orderID: validOrderID,
			mockExpect: func() {
				mockRepo.EXPECT().
					AddOrder(gomock.Any(), gomock.Any()).
					Return(model.ErrOrderExists)

				mockRepo.EXPECT().
					GetOrder(gomock.Any(), validOrderID, userID).
					Return(&model.Order{ID: validOrderID, UserID: userID}, nil)
			},
			wantErr: model.ErrOrderExists,
		},
		{
			name:    "Error - internal DB error on AddOrder",
			orderID: validOrderID,
			mockExpect: func() {
				mockRepo.EXPECT().
					AddOrder(gomock.Any(), gomock.Any()).
					Return(errors.New("connection lost"))
			},
			wantErr: errors.New("connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			err := service.Store(ctx, tt.orderID)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, model.ErrOrderWrongFormat) || errors.Is(tt.wantErr, model.ErrOrderExists) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOrderActions_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	outCh := make(chan *processors.Task)

	processTaskChannel(t, outCh)

	mockRepo := repoMocks.NewMockSQLStorageInterface(ctrl)
	mockClient := clientMocks.NewMockAccrualClienttInterface(ctrl)
	service := services.NewOrderService(mockRepo, mockClient, outCh)

	userID := "user_1"
	ctx := getCtxWithClaims(t, userID)

	tests := []struct {
		name       string
		mockExpect func()
		wantList   model.OrdersList
		wantErr    error
	}{
		{
			name: "Success - list found",
			mockExpect: func() {
				orders := model.OrdersList{
					&model.Order{ID: "1", UserID: userID},
				}
				mockRepo.EXPECT().
					ListOrders(gomock.Any(), userID).
					Return(orders, nil)
			},
			wantList: model.OrdersList{
				&model.Order{ID: "1", UserID: userID},
			},
			wantErr: nil,
		},
		{
			name: "Success - empty list (no orders)",
			mockExpect: func() {
				mockRepo.EXPECT().
					ListOrders(gomock.Any(), userID).
					Return(model.OrdersList{}, nil)
			},
			wantList: model.OrdersList{},
			wantErr:  nil,
		},
		{
			name: "Error - DB error",
			mockExpect: func() {
				mockRepo.EXPECT().
					ListOrders(gomock.Any(), userID).
					Return(nil, errors.New("db error"))
			},
			wantList: nil,
			wantErr:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			list, err := service.List(ctx)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr.Error())
				require.Nil(t, list)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantList, list)
			}
		})
	}
}
