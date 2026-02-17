package orders

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepo_UpdateOrder(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	orderNumber := "123"
	status := consts.PROCESSED
	accrual := 500

	mock.ExpectExec(regexp.QuoteMeta(updateOrderQuery())).
		WithArgs(status, accrual, orderNumber).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := NewOrderRepo(mock)
	err = repo.UpdateOrder(t.Context(), dto.AccrualCalculatorDTO{
		Order:   orderNumber,
		Status:  status,
		Accrual: accrual,
	})
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_RegisterOrder(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	orderNumber := "123"
	userUuid := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(insertOrderQuery())).
		WithArgs(orderNumber, userUuid, consts.REGISTERED, 0).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := NewOrderRepo(mock)
	ctx := context.WithValue(t.Context(), consts.UserIdKey, userUuid)
	err = repo.RegisterOrder(ctx, orderNumber)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_GetOrderInfoKeyPrefix(t *testing.T) {
	userID := "u1"
	assert.Equal(t, "u1_", getOrderInfoKeyPrefix(userID))
}

func TestOrderRepo_GetOrderInfoKey(t *testing.T) {
	userID := "u1"
	order := "123"
	assert.Equal(t, "u1_123", getOrderInfoKey(userID, order))
}

func TestOrderRepo_OrderInfoSliceToGetOrdersInfoResp(t *testing.T) {
	ts := time.Now()

	orders := []*dto.OrderInfo{
		{Number: "1", Status: consts.REGISTERED, Accrual: 0, UploadedAt: ts},
	}

	res := orderInfoSliceToGetOrdersInfoResp(orders)

	assert.Len(t, res, 1)
	assert.Equal(t, "1", res[0].Number)
	assert.Equal(t, consts.REGISTERED, res[0].Status)
	assert.True(t, res[0].UploadedAt.Equal(ts))
}

func TestOrderRepo_GetOrders_InvalidUserIDType(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	repo := NewOrderRepo(mock)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, 123)
	_, err := repo.GetOrders(ctx)

	assert.Error(t, err)
}

func TestOrderRepo_GetOrders_FromCache(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	repo := NewOrderRepo(mock)

	userID := "u1"
	ts := time.Now()

	repo.cachedOrders.Set(getOrderInfoKey(userID, "123"), &dto.OrderInfo{
		Number:     "123",
		Status:     consts.REGISTERED,
		Accrual:    0,
		UploadedAt: ts,
	})

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	res, err := repo.GetOrders(ctx)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_GetOrders_FromDB(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	repo := NewOrderRepo(mock)

	userID := "u1"
	ts := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(getOrdersQuery())).
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
				AddRow("111", consts.PROCESSED, 100, ts),
		)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	res, err := repo.GetOrders(ctx)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_GetOrders_DBQueryError(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	repo := NewOrderRepo(mock)

	userID := "u1"

	mock.ExpectQuery(regexp.QuoteMeta(getOrdersQuery())).
		WithArgs(userID).
		WillReturnError(assert.AnError)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	_, err := repo.GetOrders(ctx)

	assert.Error(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_GetOrders_ScanError(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	repo := NewOrderRepo(mock)

	userID := "u1"
	ts := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(getOrdersQuery())).
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
				AddRow("111", consts.PROCESSED, "bad", ts),
		)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	_, err := repo.GetOrders(ctx)

	assert.Error(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}
