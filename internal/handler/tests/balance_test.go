package tests

import (
	"Loyalty/configs"
	"Loyalty/internal/client"
	"Loyalty/internal/handler"
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/pashagolub/pgxmock"
	"github.com/sirupsen/logrus"
)

//====================================================================
func Test_GetBalance(t *testing.T) {
	type want struct {
		statusCode int
		balance    models.Balance
	}
	tests := []struct {
		name  string
		login string
		want  want
	}{
		{
			name:  "error",
			login: "Vasya",
			want: want{
				statusCode: 500,
				balance: models.Balance{
					Current:   0,
					Withdrawn: 0,
				},
			},
		},
		{
			name:  "ok",
			login: "Petya",
			want: want{
				statusCode: 200,
				balance: models.Balance{
					Current:   500,
					Withdrawn: 12,
				},
			},
		},
	}
	//init logger
	logger := logrus.New()

	//init config
	config := configs.NewConfigForTest()

	//init accrual client
	c := client.NewAccrualClient(logger, config.AccrualAddress)

	//init mock db connection
	mock, err := pgxmock.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer mock.Close(context.Background())

	//init main components
	r := repository.NewRepository(mock, logger)
	s := service.NewService(r, c, logger)
	h := handler.NewHandler(s, logger)

	//run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var current, withdraw uint64
			current = 50000
			withdraw = 1200
			//set mock
			urlRows := mock.NewRows([]string{"current", "withdrawn"}).AddRow(current, withdraw)

			mock.ExpectQuery("SELECT current, withdrawn FROM accounts WHERE").
				WithArgs("Petya").
				WillReturnRows(urlRows)

			h.UserLogin = tt.login

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			w := httptest.NewRecorder()

			//init router
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.GET("/api/user/balance", h.GetBalance)

			//sent request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			b, err := io.ReadAll(result.Body)
			if err != nil {
				logger.Error(err)
			}
			var balance models.Balance
			if err := json.Unmarshal(b, &balance); err != nil {
				logger.Error(err)
			}

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
			assert.Equal(t, balance, tt.want.balance)

		})
	}

}
