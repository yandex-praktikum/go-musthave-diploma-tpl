package orders

import (
	"context"
	"regexp"
	"testing"

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
	ctx := context.WithValue(t.Context(), "userID", userUuid)
	err = repo.RegisterOrder(ctx, orderNumber)
	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}
