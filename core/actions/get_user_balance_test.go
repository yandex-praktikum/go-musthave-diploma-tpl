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

func TestAuthForGetUserBalance(t *testing.T) {
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
			GetUserBalance(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.code)
		})
	}
}

func TestGetUserBalance(t *testing.T) {
	tests := []struct {
		Name                 string
		UserID               string
		GetUserBalanceResult models.GetUserBalanceDataResult
		GetUserBalanceError  error
		Code                 int
		ExpectedUserBalance  string
	}{
		{
			Name:                 "get balance with error",
			UserID:               "123",
			GetUserBalanceResult: models.GetUserBalanceDataResult{},
			GetUserBalanceError:  fmt.Errorf("some error"),
			Code:                 http.StatusBadRequest,
		},
		{
			Name:   "zero values",
			UserID: "123",
			GetUserBalanceResult: models.GetUserBalanceDataResult{
				Current:   0,
				Withdrawn: 0,
			},
			GetUserBalanceError: nil,
			Code:                http.StatusOK,
			ExpectedUserBalance: `{"current":0,"withdrawn":0}`,
		},
		{
			Name:   "non zero values",
			UserID: "123",
			GetUserBalanceResult: models.GetUserBalanceDataResult{
				Current:   42,
				Withdrawn: 117.56,
			},
			GetUserBalanceError: nil,
			Code:                http.StatusOK,
			ExpectedUserBalance: `{"current":42,"withdrawn":117.56}`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			gomock.InOrder(
				mockStore.EXPECT().GetUserBalance(gomock.Any(), models.GetUserBalanceData{UserID: test.UserID}).Return(test.GetUserBalanceResult, test.GetUserBalanceError).Times(1),
			)

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request = request.WithContext(auth.UpdateContext(request.Context(), test.UserID))

			w := httptest.NewRecorder()
			GetUserBalance(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.Code)

			b, err := io.ReadAll(res.Body)

			assert.NoError(t, err)
			assert.Equal(t, string(b), test.ExpectedUserBalance)
		})
	}
}
