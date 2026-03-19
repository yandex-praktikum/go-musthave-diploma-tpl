package orders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gmart/internal/model/auth"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stretchr/testify/assert"
)

func TestOrders_Routes_Registration(t *testing.T) {
	// 1. Инициализируем минимальное окружение (mux + huma)
	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("Orders Test API", "1.0.0")
	api := humago.New(mux, humaConfig)

	// Нам не нужен реальный репозиторий, так как Auth Middleware
	// должна прервать запрос до вызова хендлера.
	o := &Orders{}

	// Используем реальный верификатор для работы Middleware
	// (секретный ключ должен быть не менее 32 байт согласно нашей проверке в NewTokenGenerator)
	vrf, _ := auth.NewTokenVerifier(nil, []byte("test-secret-key-32-chars-long-!!!"))

	// 2. Регистрируем роуты слайса заказов
	o.RegistryRoutes(api, vrf)

	// Список ручек из ТЗ для этого слайса
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/user/orders"},
		{http.MethodGet, "/api/user/orders"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			// Создаем запрос без токена
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rr := httptest.NewRecorder()

			// Направляем запрос в мультиплексор
			mux.ServeHTTP(rr, req)

			// Проверка 1: Ручка найдена (не 404)
			assert.NotEqual(t, http.StatusNotFound, rr.Code, "Маршрут %s %s не зарегистрирован", rt.method, rt.path)

			// Проверка 2: Ручка защищена (401 Unauthorized)
			// Это подтверждает, что authMW корректно применен к этим путям
			assert.Equal(t, http.StatusUnauthorized, rr.Code, "Маршрут %s %s должен требовать авторизацию", rt.method, rt.path)
		})
	}
}
