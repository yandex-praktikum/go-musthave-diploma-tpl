package loyalty

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gmart/internal/model/auth"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stretchr/testify/assert"
)

func TestLoyalty_Routes_Registration(t *testing.T) {
	// 1. Инициализируем минимальное окружение
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))

	// Нам не нужен реальный репозиторий, так как мы не дойдем до хендлера (отсечет Auth)
	l := &Loyalty{}

	// Используем реальный верификатор, чтобы Middleware работала штатно
	vrf, _ := auth.NewTokenVerifier(nil, []byte("test-secret-key-32-chars-long-!!!"))

	// 2. Регистрируем роуты
	l.RegistryRoutes(api, vrf)

	// Список ручек, которые ОБЯЗАНЫ быть в этом слайсе по ТЗ
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/user/balance"},
		{http.MethodPost, "/api/user/balance/withdraw"},
		{http.MethodGet, "/api/user/withdrawals"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			// Если мы получили 401 — значит:
			// 1. Ручка найдена (иначе был бы 404)
			// 2. На ручке висит Auth Middleware (она проверила отсутствие токена)
			assert.Equal(t, http.StatusUnauthorized, rr.Code, "Маршрут %s %s должен требовать авторизацию", rt.method, rt.path)
			assert.NotEqual(t, http.StatusNotFound, rr.Code, "Маршрут %s %s не зарегистрирован", rt.method, rt.path)
		})
	}
}
