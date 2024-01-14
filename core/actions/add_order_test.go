package actions

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store"
	"github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store/models"
	"github.com/k-morozov/go-musthave-diploma-tpl/components/auth"
	"github.com/k-morozov/go-musthave-diploma-tpl/mocks"
)

func TestAuthForAddOrder(t *testing.T) {
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
			request := httptest.NewRequest(http.MethodPost, "/", nil)

			w := httptest.NewRecorder()
			AddOrder(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.code)
		})
	}
}

func TestAddOrder(t *testing.T) {
	tests := []struct {
		name                  string
		userID                string
		ownerOrder            string
		orderID               string
		body                  *bytes.Buffer
		code                  int
		AddOrderError         error
		GetOwnerForOrderError error
	}{
		{
			name:          "new order",
			userID:        "123",
			ownerOrder:    "123",
			orderID:       "123456",
			body:          bytes.NewBufferString(`123456`),
			code:          http.StatusAccepted,
			AddOrderError: nil,
		},
		{
			name:          "new order with error",
			userID:        "123",
			ownerOrder:    "123",
			orderID:       "123456",
			body:          bytes.NewBufferString(`123456`),
			code:          http.StatusInternalServerError,
			AddOrderError: fmt.Errorf("unknown error from store"),
		},
		{
			name:                  "new order with error",
			userID:                "123",
			ownerOrder:            "123",
			orderID:               "123456",
			body:                  bytes.NewBufferString(`123456`),
			code:                  http.StatusInternalServerError,
			AddOrderError:         store.ErrorUniqueViolation{},
			GetOwnerForOrderError: fmt.Errorf("unknown error from store"),
		},
		{
			name:                  "current user has this order yet",
			userID:                "123",
			ownerOrder:            "123",
			orderID:               "123456",
			body:                  bytes.NewBufferString(`123456`),
			code:                  http.StatusOK,
			AddOrderError:         store.ErrorUniqueViolation{},
			GetOwnerForOrderError: nil,
		},
		{
			name:                  "another user has this order yet",
			userID:                "123",
			ownerOrder:            "2",
			orderID:               "123456",
			body:                  bytes.NewBufferString(`123456`),
			code:                  http.StatusConflict,
			AddOrderError:         store.ErrorUniqueViolation{},
			GetOwnerForOrderError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			gomock.InOrder(
				mockStore.EXPECT().AddOrder(gomock.Any(), models.NewAddOrderData(test.userID, test.orderID)).Return(test.AddOrderError).Times(1),
				mockStore.EXPECT().GetOwnerForOrder(gomock.Any(), models.GetOwnerForOrderData{
					OrderID: test.orderID,
				}).Return(test.ownerOrder, test.GetOwnerForOrderError).MaxTimes(1),
			)

			request := httptest.NewRequest(http.MethodPost, "/", test.body)
			request = request.WithContext(auth.UpdateContext(request.Context(), test.userID))

			w := httptest.NewRecorder()
			AddOrder(w, request, mockStore)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, test.code)
		})
	}
}
