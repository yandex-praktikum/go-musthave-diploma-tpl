package actions

import (
	"bytes"
	"fmt"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"github.com/k-morozov/go-musthave-diploma-tpl/mocks"
)

func TestAuthForWithdraw(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{
			name: "failed auth",
			code: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			w := httptest.NewRecorder()
			Withdraw(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.code)
		})
	}
}

func TestWithdraw(t *testing.T) {
	tests := []struct {
		Name             string
		Data             models.WithdrawData
		UserID           string
		OrderID          string
		Sum              float64
		body             *bytes.Buffer
		WithdrawError    error
		Code             int
		ExpectedWithdraw string
	}{
		{
			Name:          "get balance",
			UserID:        "1",
			OrderID:       "11",
			Sum:           10.1,
			body:          bytes.NewBufferString(`{"order":"11","sum":10.1}`),
			WithdrawError: nil,
			Code:          http.StatusOK,
		},
		{
			Name:          "get balance with error",
			UserID:        "1",
			OrderID:       "11",
			Sum:           10.1,
			body:          bytes.NewBufferString(`{"order":"11","sum":10.1}`),
			WithdrawError: fmt.Errorf("some error"),
			Code:          http.StatusBadRequest,
		},
		{
			Name:          "user no money",
			UserID:        "1",
			OrderID:       "11",
			Sum:           10.1,
			body:          bytes.NewBufferString(`{"order":"11","sum":10.1}`),
			WithdrawError: store.UserNoMoney{},
			Code:          http.StatusPaymentRequired,
		},
		{
			Name:          "broken body",
			UserID:        "1",
			body:          bytes.NewBufferString(`{"orde.1}`),
			WithdrawError: nil,
			Code:          http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			gomock.InOrder(
				mockStore.EXPECT().Withdraw(gomock.Any(), models.WithdrawData{
					UserID:  test.UserID,
					OrderID: test.OrderID,
					Sum:     test.Sum,
				}).Return(test.WithdrawError).MaxTimes(1),
			)

			request := httptest.NewRequest(http.MethodPost, "/", test.body)
			request = request.WithContext(auth.UpdateContext(request.Context(), test.UserID))

			w := httptest.NewRecorder()
			Withdraw(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.Code)

			b, err := io.ReadAll(res.Body)

			assert.NoError(t, err)
			assert.Equal(t, string(b), test.ExpectedWithdraw)
		})
	}
}
