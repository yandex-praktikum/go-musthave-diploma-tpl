package withdrawals

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/utils"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawalsRepo_RegisterWithdraw_InvalidUserIDType(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, 123)
	err = repo.RegisterWithdraw(ctx, dto.WithdrawRequest{Order: "1", Sum: 10})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid userId type")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_RegisterWithdraw_OK_CacheFilled(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	userID := "u1"
	req := dto.WithdrawRequest{Order: "2377225624", Sum: 500}

	mock.ExpectExec(regexp.QuoteMeta(insertWithdrawlQuery())).
		WithArgs(req.Order, userID, req.Sum, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	err = repo.RegisterWithdraw(ctx, req)

	assert.NoError(t, err)

	_, ok := repo.cachedWithdrawals.Get(utils.GetOrderInfoKey(userID, req.Order))
	assert.True(t, ok)

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_RegisterWithdraw_DBError_NoCache(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	userID := "u1"
	req := dto.WithdrawRequest{Order: "2377225624", Sum: 500}

	mock.ExpectExec(regexp.QuoteMeta(insertWithdrawlQuery())).
		WithArgs(req.Order, userID, req.Sum, pgxmock.AnyArg()).
		WillReturnError(assert.AnError)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	err = repo.RegisterWithdraw(ctx, req)

	assert.Error(t, err)

	_, ok := repo.cachedWithdrawals.Get(utils.GetOrderInfoKey(userID, req.Order))
	assert.False(t, ok)

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_GetWithdraws_InvalidUserIDType(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, 123)
	_, err = repo.GetWithdraws(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid userId type")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_GetWithdraws_FromCache(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	userID := "u1"
	w := &dto.WithdrawInfo{
		Order:       "1",
		Sum:         10,
		ProcessedAt: time.Now(),
	}
	repo.cachedWithdrawals.Set(utils.GetOrderInfoKey(userID, w.Order), w)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	got, err := repo.GetWithdraws(ctx)

	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, w.Order, got[0].Order)

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_GetWithdraws_FromDB(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	userID := "u1"
	ts1 := time.Now().Add(-time.Minute)
	ts2 := time.Now().Add(-2 * time.Minute)

	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsQuery())).
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"order", "sum", "processed_at"}).
				AddRow("2377225624", 500, ts1).
				AddRow("2377225625", 200, ts2),
		)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	got, err := repo.GetWithdraws(ctx)

	assert.NoError(t, err)
	assert.Len(t, got, 2)

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_GetWithdraws_DBQueryError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	userID := "u1"

	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsQuery())).
		WithArgs(userID).
		WillReturnError(assert.AnError)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	_, err = repo.GetWithdraws(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to get withdrawls from db")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestWithdrawalsRepo_GetWithdraws_ScanError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	repo := NewWithdrawalsRepo(mock)

	userID := "u1"
	ts := time.Now()

	// sum ожидается числом, дадим строку чтобы Scan упал
	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsQuery())).
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"order", "sum", "processed_at"}).
				AddRow("2377225624", "bad", ts),
		)

	ctx := context.WithValue(t.Context(), consts.UserIdKey, userID)
	_, err = repo.GetWithdraws(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetWithdraws:")
	assert.Nil(t, mock.ExpectationsWereMet())
}
