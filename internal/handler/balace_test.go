package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"

	service_mocks "github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/service/mocks"
)

func TestHandler_Withdraw(t *testing.T) {
	cfglog := &logger.Config{
		LogLevel: "info",
		DevMode:  true,
		Type:     "plaintext",
	}
	cfg := &config.ConfigServer{Port: "8080"}
	log := logger.NewAppLogger(cfglog)
	tests := []struct {
		name                string
		body                []byte
		currentUserID       int
		expectedStatusCode  int
		requireGenerateMock bool
		err                 error
	}{

		{
			name:                "valid",
			body:                []byte("{\"order\":\"371449635398431\",\"sum\":100}"),
			currentUserID:       1,
			expectedStatusCode:  http.StatusOK,
			requireGenerateMock: true,
		},
		{
			name:                "Unauthorized",
			body:                []byte("{\"order\":\"371449635398431\",\"sum\":100}"),
			currentUserID:       0,
			expectedStatusCode:  http.StatusUnauthorized,
			requireGenerateMock: false,
		},
		{
			name:                "Payment Required",
			body:                []byte("{\"order\":\"371449635398431\",\"sum\":10000}"),
			currentUserID:       1,
			expectedStatusCode:  http.StatusPaymentRequired,
			requireGenerateMock: true,
			err:                 errors.New("PaymentRequired"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := service_mocks.NewMockBalance(ctrl)
			if test.requireGenerateMock {
				gomock.InOrder(
					repo.EXPECT().Withdraw(gomock.Any(), gomock.Any()).Return(test.err),
				)
			}
			handler := NewHandler(nil, nil, repo, cfg, log)

			// Setup Test Server
			// Init Endpoint
			router := gin.New()

			if test.currentUserID != 0 {
				router.Use(func(c *gin.Context) {
					c.Set("userId", test.currentUserID)
				})

			}
			router.POST("/api/user/balance/withdraw", handler.Withdraw)
			// Create Request
			writer := httptest.NewRecorder()

			request := httptest.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(test.body))
			// Make Request
			// Выполняем ОДИН запрос к серверу
			router.ServeHTTP(writer, request)

			// require
			require.Equal(t, test.expectedStatusCode, writer.Code)
		})
	}
}
