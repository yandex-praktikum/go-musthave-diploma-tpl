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

func TestAuthForGetOrders(t *testing.T) {
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
			GetOrders(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.code)
		})
	}
}

func TestGetOrders(t *testing.T) {
	tests := []struct {
		Name            string
		UserID          string
		GetOrdersResult models.GetOrdersDataResult
		GetOrdersError  error
		Code            int
		ExpectedOrders  string
	}{
		{
			Name:            "get orders with error",
			UserID:          "123",
			GetOrdersResult: models.GetOrdersDataResult{},
			GetOrdersError:  fmt.Errorf("some error"),
			Code:            http.StatusBadRequest,
		},
		{
			Name:            "no orders",
			UserID:          "123",
			GetOrdersResult: models.GetOrdersDataResult{},
			GetOrdersError:  nil,
			Code:            http.StatusNoContent,
		},
		{
			Name:   "some orders",
			UserID: "123",
			GetOrdersResult: models.GetOrdersDataResult{
				Orders: []models.OrderData{
					{
						Number:     "1",
						Status:     models.NEW,
						Accrual:    0,
						UploadedAt: "",
					},
					{
						Number:     "2",
						Status:     models.PROCESSING,
						Accrual:    0,
						UploadedAt: "",
					},
					{
						Number:     "3",
						Status:     models.PROCESSED,
						Accrual:    0,
						UploadedAt: "",
					},
				},
			},
			GetOrdersError: nil,
			Code:           http.StatusOK,
			ExpectedOrders: `[{"number":"1","status":"NEW","uploaded_at":""},{"number":"2","status":"PROCESSING","uploaded_at":""},{"number":"3","status":"PROCESSED","uploaded_at":""}]`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			gomock.InOrder(
				mockStore.EXPECT().GetOrders(gomock.Any(), models.GetOrdersData{UserID: test.UserID}).Return(test.GetOrdersResult, test.GetOrdersError).Times(1),
			)

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request = request.WithContext(auth.UpdateContext(request.Context(), test.UserID))

			w := httptest.NewRecorder()
			GetOrders(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.Code)

			b, err := io.ReadAll(res.Body)

			assert.NoError(t, err)
			assert.Equal(t, string(b), test.ExpectedOrders)
		})
	}
}
