package actions

import (
	"fmt"
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

func TestAuthForWithdrawals(t *testing.T) {
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
			Withdrawals(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.code)
		})
	}
}

func TestWithdrawals(t *testing.T) {
	tests := []struct {
		Name                string
		UserID              string
		Data                models.WithdrawsDataResult
		WithdrawalsError    error
		Code                int
		ExpectedWithdrawals string
	}{
		{
			Name:             "get balance",
			UserID:           "1",
			Data:             models.WithdrawsDataResult{},
			WithdrawalsError: nil,
			Code:             http.StatusNoContent,
		},
		{
			Name:   "simple withdrawals",
			UserID: "1",
			Data: models.WithdrawsDataResult{
				Data: []models.UserWithdraw{
					{
						OrderID:     "10",
						Sun:         10,
						ProcessedAt: "",
					},
					{
						OrderID:     "11",
						Sun:         7,
						ProcessedAt: "",
					},
				},
			},
			WithdrawalsError:    nil,
			Code:                http.StatusOK,
			ExpectedWithdrawals: `[{"order":"10","sum":10,"processed_at":""},{"order":"11","sum":7,"processed_at":""}]`,
		},
		{
			Name:             "get error from store",
			UserID:           "1",
			Data:             models.WithdrawsDataResult{},
			WithdrawalsError: fmt.Errorf("some error"),
			Code:             http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			gomock.InOrder(
				mockStore.EXPECT().Withdrawals(gomock.Any(), models.WithdrawalsData{
					UserID: test.UserID,
				}).Return(test.Data, test.WithdrawalsError).MaxTimes(1),
			)

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request = request.WithContext(auth.UpdateContext(request.Context(), test.UserID))

			w := httptest.NewRecorder()
			Withdrawals(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.Code)

			b, err := io.ReadAll(res.Body)

			assert.NoError(t, err)
			assert.Equal(t, string(b), test.ExpectedWithdrawals)
		})
	}
}
