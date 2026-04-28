package gophermart_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KalessinD/gophermart/internal/config"
	"github.com/KalessinD/gophermart/internal/gophermart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetBaseRouter(t *testing.T) {
	t.Helper()

	log := zap.NewNop()
	cfg := &config.GophermartConfig{
		ProcessingTimeout: 10 * time.Second,
	}

	router := gophermart.GetBaseRouter(cfg, log)

	assert.NotNil(t, router, "Router should not be nil")
}

func TestNewRouter(t *testing.T) {
	log := zap.NewNop()
	cfg := &config.GophermartConfig{
		ProcessingTimeout: 10 * time.Second,
		EncryptionKey:     "test-secret-key-for-jwt",
	}

	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler, err := gophermart.NewRouter(cfg, log, db)
	require.NoError(t, err)
	assert.NotNil(t, handler)

	t.Run("check registered routes", func(t *testing.T) {
		// Список маршрутов, которые должны быть зарегистрированы
		routes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/user/login"},
			{"POST", "/api/user/register"},
			{"POST", "/api/user/orders"},
			{"GET", "/api/user/orders"},
			{"GET", "/api/user/balance"},
			{"POST", "/api/user/balance/withdraw"},
			{"GET", "/api/user/withdrawals"},
		}

		for _, route := range routes {
			t.Run(route.method+" "+route.path, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)

				// Мы не проверяем логику хендлеров (это юнит-тесты хендлеров),
				// мы проверяем, что роутер нашел маршрут.
				// 404 означает, что маршрут не зарегистрирован.
				// Любой другой код (400, 401, 500) означает, что маршрут найден.
				assert.NotEqual(t, http.StatusNotFound, rec.Code, "Route should be registered")
			})
		}
	})

	t.Run("check unauthorized access to protected routes", func(t *testing.T) {
		// Проверяем, что защищенные роуты возвращают 401 без токена
		protectedRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/user/orders"},
			{"GET", "/api/user/orders"},
			{"GET", "/api/user/balance"},
			{"POST", "/api/user/balance/withdraw"},
			{"GET", "/api/user/withdrawals"},
		}

		for _, route := range protectedRoutes {
			t.Run(route.method+" "+route.path, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)

				// Ожидаем 401, так как нет токена
				assert.Equal(t, http.StatusUnauthorized, rec.Code, "Protected route should return 401")
			})
		}
	})

	// Проверка ожиданий мока
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestPsqlConnect_Failure тестирует сценарий ошибки подключения.
// Успешное подключение юнит-тестом не покрывается, так как требует реальной БД.
func TestPsqlConnect_Failure(t *testing.T) {
	log := zap.NewNop()
	ctx := context.Background()

	// Несуществующий DSN
	dsn := "host=localhost port=9999 user=invalid password=invalid dbname=invalid sslmode=disable"

	// sql.Open может не вернуть ошибку сразу, а Ping вернет
	db, err := gophermart.PsqlConnect(ctx, dsn, log)
	assert.Error(t, err, "Should return error for invalid DSN")
	assert.Nil(t, db)
}
