package dto

import (
	"testing"
	"time"

	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/stretchr/testify/require"
)

func TestNewOrderInfo(t *testing.T) {
	order := NewOrderInfo("123", "user-1")

	require.Equal(t, "123", order.Number)
	require.Equal(t, "user-1", order.GetUserId())
	require.Equal(t, consts.REGISTERED, order.Status)
	require.Equal(t, 0, order.Accrual)
	require.False(t, order.UploadedAt.IsZero())
}

func TestOrderInfo_ToGetOrdersInfoResp(t *testing.T) {
	now := time.Now()
	order := &OrderInfo{
		Number:     "123",
		Status:     "PROCESSED",
		UploadedAt: now,
	}

	resp := order.ToGetOrdersInfoResp()

	require.Equal(t, "123", resp.Number)
	require.Equal(t, "PROCESSED", resp.Status)
	require.Equal(t, now, resp.UploadedAt)
}

func TestOrderInfo_IsEqual(t *testing.T) {
	order := &OrderInfo{Status: "PROCESSED"}
	dto := &AccrualCalculatorDTO{Status: "PROCESSED"}

	require.True(t, order.IsEqual(dto))

	dto.Status = "INVALID"
	require.False(t, order.IsEqual(dto))
}

func TestOrderInfo_Update(t *testing.T) {
	order := &OrderInfo{
		Status:  "NEW",
		Accrual: 0,
	}

	dto := &AccrualCalculatorDTO{
		Status:  "PROCESSED",
		Accrual: 500,
	}

	order.Update(dto)

	require.Equal(t, "PROCESSED", order.Status)
	require.Equal(t, 500, order.Accrual)
}

func TestOrderInfo_GetUserId(t *testing.T) {
	order := &OrderInfo{userID: "user-42"}
	require.Equal(t, "user-42", order.GetUserId())
}

func TestUserCredential_ToUserData(t *testing.T) {
	cred := &UserCredential{
		Login:    "user1",
		Password: "pass123",
	}

	uuid := "3f3b0e7e-5f8a-4d2a-9b66-1c8a2b6e1f7a"

	user := cred.ToUserData(uuid)

	require.Equal(t, uuid, user.Uuid)
	require.Equal(t, "user1", user.Login)
	require.Equal(t, "pass123", user.Password)
}

func TestNewWithdrawInfo(t *testing.T) {
	req := WithdrawRequest{
		Order: "123",
		Sum:   500,
	}

	w := NewWithdrawInfo(req)

	require.Equal(t, "123", w.Order)
	require.Equal(t, 500, w.Sum)
	require.False(t, w.ProcessedAt.IsZero())
}

func TestAccrualCalculatorDTO_IsEqual(t *testing.T) {
	a := &AccrualCalculatorDTO{
		Order:   "1",
		Status:  "PROCESSED",
		Accrual: 100,
	}

	b := AccrualCalculatorDTO{
		Order:   "1",
		Status:  "PROCESSED",
		Accrual: 100,
	}

	require.True(t, a.IsEqual(b))

	b.Status = "INVALID"
	require.False(t, a.IsEqual(b))
}

func TestAccrualCalculatorDTO_UserId(t *testing.T) {
	dto := &AccrualCalculatorDTO{}

	dto.AddUserId("user-1")

	require.Equal(t, "user-1", dto.GetUserId())
}
