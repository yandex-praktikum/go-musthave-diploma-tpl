package accrual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalConst "github.com/Raime-34/gophermart.git/internal/accrual/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_composeBaseUrl(t *testing.T) {
	targetServiceUrl := "localhost:8080"
	expectedComposedUrl := "localhost:8080/api/orders/%v"

	composedUrl := composeBaseUrl(targetServiceUrl)
	assert.Equal(t, expectedComposedUrl, composedUrl)
}

func TestAccrualCalculator_getOrderState(t *testing.T) {
	orderInfo := dto.AccrualCalculatorDTO{
		Order:   "123",
		Status:  consts.PROCESSED,
		Accrual: 10,
	}
	b, _ := json.Marshal(orderInfo)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer ts.Close()

	calculator := NewAccrualCalculator(ts.URL)
	gotedOrderInfo, err := calculator.getOrderState(orderInfo.Order)
	assert.NotNil(t, orderInfo)
	assert.Nil(t, err)
	assert.Equal(t, orderInfo.Order, gotedOrderInfo.Order)
	assert.Equal(t, orderInfo.Status, gotedOrderInfo.Status)
	assert.Equal(t, orderInfo.Accrual, gotedOrderInfo.Accrual)
}

func TestAccrualCalculator_AddToMonitoring(t *testing.T) {
	calculator := NewAccrualCalculator("")
	userID := uuid.New().String()
	calculator.AddToMonitoring("123", userID)

	assert.Equal(t, 1, len(calculator.orderStates))
}

func TestAccrualCalculator_getOrderState_NoContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)

	got, err := c.getOrderState("123")

	assert.Nil(t, got)
	assert.ErrorIs(t, err, internalConst.ErrNotRegistered)
}

func TestAccrualCalculator_getOrderState_TooManyRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)

	got, err := c.getOrderState("123")

	assert.Nil(t, got)
	assert.ErrorIs(t, err, internalConst.ErrToManyRequest)
}

func TestAccrualCalculator_getOrderState_InternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)

	got, err := c.getOrderState("123")

	assert.Nil(t, got)
	assert.ErrorIs(t, err, internalConst.ErrInternal)
}

func TestAccrualCalculator_getOrderState_UnknownStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)

	got, err := c.getOrderState("123")

	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown status")
}

func TestAccrualCalculator_getOrderState_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order":`)) // битый JSON
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)

	got, err := c.getOrderState("123")

	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to decode")
}

func TestAccrualCalculator_getOrderState_RequestError_ReturnsNilNil(t *testing.T) {
	// порт 0 гарантирует connection error
	c := NewAccrualCalculator("http://127.0.0.1:0")

	got, err := c.getOrderState("123")

	assert.Nil(t, err)
	assert.Nil(t, got)
}

func TestAccrualCalculator_StartMonitoring_SendsUpdateOnStatusChange(t *testing.T) {
	orderInfo := dto.AccrualCalculatorDTO{
		Order:   "123",
		Status:  consts.PROCESSED,
		Accrual: 50,
	}
	b, _ := json.Marshal(orderInfo)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)
	userID := uuid.New().String()
	c.AddToMonitoring("123", userID)

	ctx, cancel := context.WithCancel(t.Context())
	ch := c.StartMonitoring(ctx)

	select {
	case got := <-ch:
		assert.NotNil(t, got)
		assert.Equal(t, "123", got.Order)
		assert.Equal(t, consts.PROCESSED, got.Status)
		assert.Equal(t, 50, got.Accrual)
		cancel()
	case <-time.After(500 * time.Millisecond):
		cancel()
		t.Fatal("no update received")
	}
}

func TestAccrualCalculator_StartMonitoring_CtxCancel_ClosesChannel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order":"1","status":"PROCESSED","accrual":10}`))
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)

	ctx, cancel := context.WithCancel(t.Context())
	ch := c.StartMonitoring(ctx)

	cancel()

	select {
	case _, ok := <-ch:
		assert.False(t, ok)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("channel not closed")
	}
}

func TestAccrualCalculator_StartMonitoring_ErrNotRegistered_DeletesOrder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewAccrualCalculator(ts.URL)
	userID := uuid.New().String()
	c.AddToMonitoring("123", userID)

	ctx, cancel := context.WithCancel(t.Context())
	_ = c.StartMonitoring(ctx)

	assert.Eventually(t, func() bool {
		c.mu.RLock()
		defer c.mu.RUnlock()
		_, ok := c.orderStates["123"]
		return !ok
	}, 500*time.Millisecond, 10*time.Millisecond)

	cancel()
}
