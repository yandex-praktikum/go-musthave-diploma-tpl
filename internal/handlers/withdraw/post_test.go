package withdraw

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/logger"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/orders"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerPost(t *testing.T) {
	tests := []struct {
		name           string
		login          string
		requestBody    RequestBody
		whenBodyBad    bool
		responseError  error
		expectedStatus int
	}{
		{
			name:  "Successful post",
			login: "test",
			requestBody: RequestBody{
				Order: "22664155",
				Sum:   55.5,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Invalid body",
			login: "test",
			requestBody: RequestBody{
				Order: "22664155",
				Sum:   55.5,
			},
			whenBodyBad:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Invalid order",
			login: "test",
			requestBody: RequestBody{
				Order: "226641551",
				Sum:   55.5,
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "Invalid login",
			login: "",
			requestBody: RequestBody{
				Order: "22664155",
				Sum:   55.5,
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "order is not in the database",
			login: "test",
			requestBody: RequestBody{
				Order: "22664155",
				Sum:   55.5,
			},
			responseError:  customerrors.ErrNotData,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "don't have enough bonuses",
			login: "test",
			requestBody: RequestBody{
				Order: "22664155",
				Sum:   55.5,
			},
			responseError:  customerrors.ErrNotEnoughBonuses,
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:  "check status 500",
			login: "test",
			requestBody: RequestBody{
				Order: "22664155",
				Sum:   55.5,
			},
			responseError:  errors.New("check status 500"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			loger := logger.NewLogger()
			ctx := context.Background()

			repo := orders.NewMockStorage(ctrl)
			repo.EXPECT().CheckWriteOffOfFunds(ctx, tt.login, tt.requestBody.Order, tt.requestBody.Sum, gomock.Any()).Return(tt.responseError).AnyTimes()

			serv := orders.NewService(repo, loger)
			handler := NewHandler(ctx, serv, loger)

			req, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			req = req.WithContext(context.WithValue(req.Context(), middleware.LoginContentKey, tt.login))

			if tt.whenBodyBad {
				req.Body = io.NopCloser(bytes.NewBufferString("incorrect body"))
			} else {
				req.Body = jsonReader(tt.requestBody)
			}

			w := httptest.NewRecorder()

			handler.Post(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func jsonReader(v interface{}) io.ReadCloser {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return io.NopCloser(bytes.NewBuffer(b))
}
