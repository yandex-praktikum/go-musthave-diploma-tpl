package order

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/models"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerGet(t *testing.T) {
	tests := []struct {
		name           string
		login          string
		order          []*models.OrdersUser
		responseError  error
		expectedStatus int
	}{
		{
			name:  "Successful get",
			login: "test",
			order: []*models.OrdersUser{
				{
					Number: "1",
				},
				{
					Number: "2",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User not authenticated",
			login:          "",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "No data to answer",
			login: "test",
			order: []*models.OrdersUser{
				{
					Number: "1",
				},
				{
					Number: "2",
				},
			},
			responseError:  customerrors.ErrNotFound,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Cannot loading order",
			login:          "test",
			order:          []*models.OrdersUser{},
			responseError:  errors.New("cannot loading order"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Len data 0",
			login:          "test",
			order:          []*models.OrdersUser{},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			loger := logger.NewLogger()

			repo := orders.NewMockStorage(ctrl)
			repo.EXPECT().GetAllUserOrders(tt.login).Return(tt.order, tt.responseError).AnyTimes()

			serv := orders.NewService(repo, loger)

			handler := NewHandler(serv, loger)

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req = req.WithContext(context.WithValue(req.Context(), middleware.LoginContentKey, tt.login))
			w := httptest.NewRecorder()

			handler.Get(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected expectedStatus code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
