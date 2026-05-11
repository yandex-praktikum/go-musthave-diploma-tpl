//go:build e2e

package gophermart_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/KalessinD/gophermart/tests/e2e/fixtures"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// комбинирует PostgresSuite и AccrualSuite
type OrderAccrualE2ETestSuite struct {
	fixtures.PostgresSuite
	accrual    fixtures.AccrualSuite
	server     *httptest.Server
	accrualURL string
}

func (s *OrderAccrualE2ETestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()

	err := s.accrual.SetupAccrual(s.Ctx)
	require.NoError(s.T(), err, "Failed to start accrual container")
	s.accrualURL = s.accrual.AccrualURL

	connStr, err := s.Container.ConnectionString(s.Ctx, "sslmode=disable")
	require.NoError(s.T(), err, "Failed to get connection string")

	cfg := fixtures.NewTestConfig(connStr, s.T().TempDir())
	cfg.AccrualAddress = s.accrualURL // переопределяем адрес accrual в конфиге

	router, err := gophermart.NewRouter(s.Ctx, cfg, s.Log, s.DB)
	require.NoError(s.T(), err, "Failed to create router")

	s.server = httptest.NewServer(router)
}

func (s *OrderAccrualE2ETestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	s.accrual.TearDownAccrual(s.Ctx)
	s.PostgresSuite.TearDownSuite()
}

func (s *OrderAccrualE2ETestSuite) SetupTest() {
	s.PostgresSuite.SetupTest([]string{
		"gophermart.users",
		"gophermart.orders",
		"gophermart.withdrawals",
	})
}

func TestOrderAccrualE2ETestSuite(t *testing.T) {
	suite.Run(t, new(OrderAccrualE2ETestSuite))
}

// TestOrderProcessing_Success проверяет полный цикл обработки заказа с начислением баллов
func (s *OrderAccrualE2ETestSuite) TestOrderProcessing_Success() {
	// Подготовка данных в Accrual ---
	ctx := context.Background()
	goodMatch := "TestItem"
	reward := float64(500)
	rewardType := "pt" // фиксированные баллы

	// Регистрируем товар в accrual
	err := s.accrual.RegisterGood(ctx, goodMatch, reward, rewardType)
	require.NoError(s.T(), err, "Failed to register good in accrual")

	// Регистрируем заказ в accrual
	orderNumber := "12345678903" // Валидный по Луну
	goods := []map[string]any{
		{
			"description": "Some item", // Не совпадает
			"price":       1000,
		},
		{
			"description": goodMatch, // Совпадает с паттерном
			"price":       200,
		},
	}
	err = s.accrual.RegisterOrder(ctx, orderNumber, goods)
	require.NoError(s.T(), err, "Failed to register order in accrual")

	// Действия пользователя в Gophermart ---

	// Авторизация
	cookie := fixtures.AuthUser(s.T(), s.server, "user_accrual", "password123")

	// Загрузка заказа в Gophermart
	req, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	require.Equal(s.T(), http.StatusAccepted, resp.StatusCode, "Order should be accepted")

	// Ожидание обработки (воркер должен опросить accrual и обновить статус)
	var listResp *http.Response
	assert.Eventually(s.T(), func() bool {
		reqList, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/orders", nil)
		require.NoError(s.T(), err, "Failed to list user's orders")

		reqList.AddCookie(cookie)
		listResp, err = client.Do(reqList)
		if err != nil || listResp.StatusCode != http.StatusOK {
			return false
		}
		defer listResp.Body.Close()

		body, err := io.ReadAll(listResp.Body)
		require.NoError(s.T(), err, "Failed to read response body")

		// Ищем статус PROCESSED и наличие accrual
		return strings.Contains(string(body), `"status":"PROCESSED"`) && strings.Contains(string(body), `"accrual":500`)
	}, 5*time.Second, 200*time.Millisecond, "Order should be processed and accrual should be set")

	// Проверка баланса
	reqBalance, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/balance", nil)
	require.NoError(s.T(), err, "Failed to check user's balance")

	reqBalance.AddCookie(cookie)
	respBalance, err := client.Do(reqBalance)
	require.NoError(s.T(), err)
	defer respBalance.Body.Close()

	require.Equal(s.T(), http.StatusOK, respBalance.StatusCode)
	balanceBody, err := io.ReadAll(respBalance.Body)
	require.NoError(s.T(), err, "Failed to read response body")

	// Баланс должен быть 500
	assert.JSONEq(s.T(), `{"current":500,"withdrawn":0}`, string(balanceBody))
}

// TestOrderProcessing_NoAccrual проверяет случай, когда заказ обработан, но баллов не начислено
// (товары в заказе не совпали с правилами вознаграждений)
func (s *OrderAccrualE2ETestSuite) TestOrderProcessing_NoAccrual() {
	ctx := context.Background()

	// Используем валидный по Луну номер заказа, отличный от успешного теста.
	// Например: 49927398716 (стандартный тестовый номер).
	orderNumber := "49927398716"

	// Регистрируем заказ в accrual с товарами, которые не дадут награды
	// (так как мы не регистрировали правила rewards для этих товаров)
	goods := []map[string]interface{}{
		{"description": "Some random item", "price": 100},
	}
	err := s.accrual.RegisterOrder(ctx, orderNumber, goods)
	require.NoError(s.T(), err, "Failed to register order in accrual")

	// Авторизация
	cookie := fixtures.AuthUser(s.T(), s.server, "user_no_accrual", "password")
	client := &http.Client{}

	// Загрузка заказа
	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", strings.NewReader(orderNumber))
	require.NoError(s.T(), err, "Failed to create request")

	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "Failed to process the request")

	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	require.Equal(s.T(), http.StatusAccepted, resp.StatusCode)

	// Ожидание обработки: статус должен стать PROCESSED, но accrual не должно быть (или 0)
	assert.Eventually(s.T(), func() bool {
		reqList, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/orders", nil)
		require.NoError(s.T(), err, "Failed to list user's orders")

		reqList.AddCookie(cookie)
		listResp, err := client.Do(reqList)
		if err != nil || listResp.StatusCode != http.StatusOK {
			return false
		}
		defer listResp.Body.Close()
		body, err := io.ReadAll(listResp.Body)
		require.NoError(s.T(), err, "Failed to read response body")

		return strings.Contains(string(body), `"status":"PROCESSED"`) && !strings.Contains(string(body), `"accrual":500`)
	}, 5*time.Second, 200*time.Millisecond, "Order should be processed with 0 accrual")

	// Проверяем баланс (должен быть 0)
	reqBalance, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/balance", nil)
	require.NoError(s.T(), err, "Failed to get user's balance")

	reqBalance.AddCookie(cookie)
	respBalance, err := client.Do(reqBalance)
	require.NoError(s.T(), err, "Failed to make a request")

	defer respBalance.Body.Close()

	balanceBody, err := io.ReadAll(respBalance.Body)
	require.NoError(s.T(), err, "Failed to read a response")
	assert.JSONEq(s.T(), `{"current":0,"withdrawn":0}`, string(balanceBody))
}
