package tests

import (
	"Loyalty/configs"
	"Loyalty/internal/client"
	"Loyalty/internal/handler"
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	"Loyalty/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/pashagolub/pgxmock"
	"github.com/sirupsen/logrus"
)

//====================================================================
func Test_PostOrder(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name  string
		login string
		order models.Order
		want  want
	}{
		{
			name:  "notValidNumber",
			login: "Vasya",
			order: models.Order{Number: "15232167"},
			want:  want{statusCode: 422},
		},
		{
			name:  "ok",
			login: "Petya",
			order: models.Order{
				Number:  "123455",
				Status:  "NEW",
				Accrual: 0,
			},
			want: want{statusCode: 202},
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

			h.UserLogin = tt.login

			//set mock
			Rows := mock.NewRows([]string{"uploaded_at", "login"}).
				AddRow(time.Now(), "Petya")

			mock.ExpectQuery("INSERT INTO orders").
				WillReturnRows(Rows)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tt.order.Number))
			w := httptest.NewRecorder()

			//init router
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.POST("/api/user/orders", h.SaveOrder)

			//sent request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
		})
	}

}

//====================================================================
func Test_GetOrders(t *testing.T) {
	type want struct {
		statusCode int
		orders     []models.OrderDTO
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
				orders:     []models.OrderDTO{},
			},
		},
		{
			name:  "ok",
			login: "Petya",
			want: want{
				statusCode: 200,
				orders: []models.OrderDTO{
					{
						Number:     "9278923470",
						Status:     "PROCESSED",
						Accrual:    500,
						UploadedAt: time.Date(2022, 03, 02, 12, 5, 10, 0, time.UTC),
					},
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

			//set mock
			Rows := mock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
				AddRow("9278923470", "PROCESSED", 50000, time.Date(2022, 03, 02, 12, 5, 10, 0, time.UTC))

			mock.ExpectQuery("SELECT number, status, accrual, uploaded_at FROM orders WHERE").
				WithArgs("Petya").
				WillReturnRows(Rows)

			h.UserLogin = tt.login

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			w := httptest.NewRecorder()

			//init router
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.GET("/api/user/orders", h.GetOrders)

			//sent request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			b, err := io.ReadAll(result.Body)
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Info(string(b))
			var orders []models.OrderDTO
			if err := json.Unmarshal(b, &orders); err != nil {
				logger.Error(err)
				return
			}
			logger.Info(orders)

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
			assert.Equal(t, orders, tt.want.orders)

		})
	}

}
