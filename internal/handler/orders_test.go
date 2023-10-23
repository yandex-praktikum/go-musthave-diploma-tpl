package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	repo_mocks "github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository/mocks"
)

func TestHandler_PostOrder(t *testing.T) {
	cfglog := &logger.Config{
		LogLevel: "info",
		DevMode:  true,
		Type:     "plaintext",
	}
	cfg := &config.ConfigServer{Port: "8080"}
	log := logger.NewAppLogger(cfglog)
	tests := []struct {
		name               string
		body               []byte
		currentUserID      int
		statusOrder        string
		expectedStatusCode int
		updatedDate        time.Time
		requireGenerateCU  bool
	}{
		{
			name:               "valid",
			body:               []byte("371449635398431"),
			currentUserID:      1,
			statusOrder:        "New",
			expectedStatusCode: http.StatusAccepted,
			requireGenerateCU:  true,
		},
		{
			name:               "was create earlier",
			body:               []byte("371449635398431"),
			currentUserID:      1,
			statusOrder:        "New",
			expectedStatusCode: http.StatusOK,
			updatedDate:        time.Now(),
			requireGenerateCU:  true,
		},
		{
			name:               "luhn's check",
			body:               []byte("37144963539843111"),
			currentUserID:      1,
			statusOrder:        "New",
			expectedStatusCode: http.StatusUnprocessableEntity,
			updatedDate:        time.Now(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repo_mocks.NewMockOrders(ctrl)

			if test.requireGenerateCU {
				gomock.InOrder(
					repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.currentUserID, test.updatedDate, nil),
				)
			}

			handler := NewHandler(nil, repo, nil, cfg, log)

			// Setup Test Server
			// Init Endpoint
			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Set("userId", 1)
			})
			router.POST("/orders", handler.PostOrder)
			// Create Request
			writer := httptest.NewRecorder()

			request := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(test.body))

			// Make Request
			// Выполняем ОДИН запрос к серверу
			router.ServeHTTP(writer, request)

			// require

			require.Equal(t, test.expectedStatusCode, writer.Code)
		})
	}
}
