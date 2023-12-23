package handler

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/config"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/logger"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	service_mocks "github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/service/mocks"
)

func generatePasswordHash(password, salt string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
func TestHandler_SingUp(t *testing.T) {
	cfglog := &logger.Config{
		LogLevel: "info",
		DevMode:  true,
		Type:     "plaintext",
	}
	cfg := &config.ConfigServer{Port: "8080"}
	log := logger.NewAppLogger(cfglog)

	tests := []struct {
		name                 string
		inputBody            string
		inputUser            models.User
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:      "Ok",
			inputBody: `{"login": "username", "password": "qwerty"}`,
			inputUser: models.User{
				Login:    "username",
				Password: "qwerty",
			},
			err:                  nil,
			expectedStatusCode:   200,
			requireGenerateToken: true,
		},
		{
			name:      "Wrong Input",
			inputBody: `{"user": "username", "password": "qwerty"}`,
			inputUser: models.User{
				Login:    "username",
				Password: "qwerty",
			},
			err:                  nil,
			expectedStatusCode:   400,
			requireGenerateToken: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := service_mocks.NewMockAutorisation(ctrl)

			if test.requireGenerateToken {
				gomock.InOrder(
					repo.EXPECT().CreateUser(test.inputUser).Return(1, test.err),
					repo.EXPECT().GenerateToken(test.inputUser.Login, test.inputUser.Password).Return(generatePasswordHash(test.inputUser.Password, "salt"), nil),
				)
			}

			handler := NewHandler(repo, nil, nil, cfg, log)

			// Setup Test Server
			// Init Endpoint
			router := gin.New()
			router.POST("/sign-up", handler.SingUp)

			// Create Request
			writer := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/sign-up",
				bytes.NewBufferString(test.inputBody))

			// Make Request
			// Выполняем ОДИН запрос к серверу
			router.ServeHTTP(writer, request)

			header := writer.Header()["Authorization"]
			// require
			require.Equal(t, writer.Code, test.expectedStatusCode)
			require.NotEqual(t, header, "")

		})
	}
}
