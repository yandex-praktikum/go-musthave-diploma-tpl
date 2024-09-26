package balance

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
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
		getBalanceUser *models.Balance
		requestError   error
		expectedStatus int
	}{
		{
			name:  "Successful get balance",
			login: "test",
			getBalanceUser: &models.Balance{
				Current:  555.5,
				Withdraw: 50,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Invalid login",
			login: "",
			getBalanceUser: &models.Balance{
				Current:  555.5,
				Withdraw: 50,
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "Error get balance",
			login: "test",
			getBalanceUser: &models.Balance{
				Current:  555.5,
				Withdraw: 50,
			},
			requestError:   errors.New("not get balance"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			ctx := context.Background()
			loger := logger.NewLogger()

			repo := orders.NewMockStorage(ctrl)
			repo.EXPECT().GetBalanceUser(tt.login).Return(tt.getBalanceUser, tt.requestError).AnyTimes()

			serv := orders.NewService(repo, loger)
			handler := NewHandler(ctx, serv, loger)

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
