package accrual

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
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
	calculator.AddToMonitoring("123")

	assert.Equal(t, 1, len(calculator.orderStates))
}
