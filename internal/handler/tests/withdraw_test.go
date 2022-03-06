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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/pashagolub/pgxmock"
	"github.com/sirupsen/logrus"
)

//====================================================================
func Test_PostWithdraw(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name     string
		login    string
		current  uint64
		withrawn uint64
		withdraw models.WithdrawalDTO
		want     want
	}{
		{
			name:     "NotCorrectNumber",
			login:    "Petya",
			current:  0,
			withrawn: 0,
			withdraw: models.WithdrawalDTO{
				Order: "654321",
				Sum:   123,
			},
			want: want{statusCode: 422},
		},
		{
			name:     "Ok",
			login:    "Petya",
			current:  100000,
			withrawn: 0,
			withdraw: models.WithdrawalDTO{
				Order: "123455",
				Sum:   123,
			},
			want: want{statusCode: 200},
		},
		{
			name:     "NotEnoughBonuses",
			login:    "Petya",
			current:  100,
			withrawn: 0,
			withdraw: models.WithdrawalDTO{
				Order: "123455",
				Sum:   123,
			},
			want: want{statusCode: 402},
		},
	}
	//init logger
	logger := logrus.New()

	//init config
	config := configs.NewConfigForTest()

	//init accrual client
	c := client.NewAccrualClient(logger, config.AccrualAddress)

	//run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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

			h.UserLogin = tt.login

			balanceRow := mock.NewRows([]string{"current", "withdrawn"}).AddRow(tt.current, tt.withrawn)
			idRow := mock.NewRows([]string{"id"}).AddRow(1)

			orderRow := mock.NewRows([]string{"uploaded_at", "login"})

			mock.ExpectQuery("SELECT current, withdrawn	FROM accounts").
				WithArgs(tt.login).
				WillReturnRows(balanceRow)

			mock.ExpectQuery("INSERT INTO orders").WillReturnRows(orderRow.AddRow(time.Now(), tt.login))
			mock.ExpectBegin()
			mock.ExpectQuery("INSERT INTO withdrawals").WillReturnRows(idRow)

			mock.ExpectExec("UPDATE accounts SET current").WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			mock.ExpectCommit()

			body, err := json.Marshal(tt.withdraw)
			if err != nil {
				return
			}
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
			w := httptest.NewRecorder()

			//init router
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.POST("/api/user/balance/withdraw", h.Withdraw)

			//sent request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
		})
	}

}
