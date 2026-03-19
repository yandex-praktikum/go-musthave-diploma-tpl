package orders

import (
	"context"
	"errors"
	"testing"
	"time"

	"gmart/internal/domain"
	"gmart/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOrders_Upload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockOrdersRepoIFace(ctrl)
	svc := &Orders{ordersRepo: mockRepo}

	userID := domain.UserID(1)
	orderNum := domain.OrderNumber("12345678903") // Валидный по Луну (предположим)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			Upload(gomock.Any(), userID, orderNum).
			Return(nil)

		err := svc.Upload(context.Background(), userID, orderNum)
		assert.NoError(t, err)
	})

	t.Run("Conflict - Different User", func(t *testing.T) {
		mockRepo.EXPECT().
			Upload(gomock.Any(), userID, orderNum).
			Return(ErrOrderConflict)

		err := svc.Upload(context.Background(), userID, orderNum)
		assert.ErrorIs(t, err, ErrOrderConflict)
	})

	t.Run("Already Uploaded - Same User", func(t *testing.T) {
		mockRepo.EXPECT().
			Upload(gomock.Any(), userID, orderNum).
			Return(ErrOrderAlreadyUploaded)

		err := svc.Upload(context.Background(), userID, orderNum)
		assert.ErrorIs(t, err, ErrOrderAlreadyUploaded)
	})
}

func TestOrders_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockOrdersRepoIFace(ctrl)
	svc := &Orders{ordersRepo: mockRepo}

	userID := domain.UserID(1)

	t.Run("Success", func(t *testing.T) {
		expectedItems := []dto.OrderItem{
			{
				Number:     "123",
				Status:     domain.OrderStatus("NEW"),
				Accrual:    500,
				UploadedAt: time.Now(),
			},
		}

		mockRepo.EXPECT().
			List(gomock.Any(), userID).
			Return(expectedItems, nil)

		res, err := svc.List(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, domain.OrderStatus("NEW"), res[0].Status)
	})

	t.Run("Empty List", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), userID).
			Return([]dto.OrderItem{}, nil)

		res, err := svc.List(context.Background(), userID)
		assert.ErrorIs(t, err, ErrOrderListIsEmpty)
		assert.Nil(t, res)
	})

	t.Run("Repo Error", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		res, err := svc.List(context.Background(), userID)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
