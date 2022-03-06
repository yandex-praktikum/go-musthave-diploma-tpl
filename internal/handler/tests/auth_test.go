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
func Test_Register(t *testing.T) {

	type want struct {
		statusCode int
	}

	tests := []struct {
		name string
		user models.User
		want want
	}{
		{
			name: "badRequest",
			user: models.User{
				Login:    "/<>",
				Password: "sdf",
			},
			want: want{statusCode: 400},
		},
		{
			name: "loginAlreadyExist",
			user: models.User{
				Login:    "Petya",
				Password: "sddfg56456dfgj3456f",
			},
			want: want{statusCode: 409},
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
			var number uint64 = 12451234234
			//set mock
			idRow := mock.NewRows([]string{"id"}).AddRow(1)

			mock.ExpectQuery("INSERT INTO accounts").WillReturnRows(idRow)

			userRow := mock.NewRows([]string{"number"}).AddRow(number)

			mock.ExpectQuery("INSERT INTO users").WillReturnRows(userRow)

			user, err := json.Marshal(tt.user)
			if err != nil {
				logger.Error(err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(user))
			w := httptest.NewRecorder()

			//init router
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.POST("/api/user/register", h.SignIn)

			//sent request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
		})
	}

}

//====================================================================
func Test_Login(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name string
		user models.User
		want want
	}{
		{
			name: "badRequest",
			user: models.User{
				Login:    "/<>",
				Password: "",
			},
			want: want{statusCode: 400},
		},
		{
			name: "ok",
			user: models.User{
				Login:    "Petya",
				Password: "sddfg56456dfgj3456f",
			},
			want: want{statusCode: 200},
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
			var number uint64 = 12451234234
			//set mock

			userRow := mock.NewRows([]string{"number"}).AddRow(number)

			mock.ExpectQuery("SELECT number FROM users").WillReturnRows(userRow)

			user, err := json.Marshal(tt.user)
			if err != nil {
				logger.Error(err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(user))
			w := httptest.NewRecorder()

			//init router
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.POST("/api/user/login", h.SignUp)

			//sent request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
		})
	}
}
