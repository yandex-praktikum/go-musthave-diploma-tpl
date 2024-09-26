package withdraw

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
	"time"
)

func TestHandlerGet(t *testing.T) {
	var sum float32 = 55.5
	var now = time.Now()

	tests := []struct {
		name            string
		login           string
		responseError   error
		withdrawalsBody []*models.Withdrawals
		expectedStatus  int
	}{
		{
			name:  "Successful get withdraw",
			login: "test",
			withdrawalsBody: []*models.Withdrawals{
				{Order: "22222",
					Sum:         &sum,
					ProcessedAt: &now},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Bad login",
			login: "",
			withdrawalsBody: []*models.Withdrawals{
				{Order: "22222",
					Sum:         &sum,
					ProcessedAt: &now},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "Bad get withdrawals",
			login: "test",
			withdrawalsBody: []*models.Withdrawals{
				{Order: "22222",
					Sum:         &sum,
					ProcessedAt: &now},
			},
			responseError:  errors.New("bad get withdrawals"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:            "there is not a single write-off",
			login:           "test",
			withdrawalsBody: []*models.Withdrawals{},
			expectedStatus:  http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			loger := logger.NewLogger()
			ctx := context.Background()

			repo := orders.NewMockStorage(ctrl)
			repo.EXPECT().GetWithdrawals(ctx, tt.login).Return(tt.withdrawalsBody, tt.responseError).AnyTimes()

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
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
