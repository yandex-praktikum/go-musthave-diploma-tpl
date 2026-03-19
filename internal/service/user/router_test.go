package user

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stretchr/testify/assert"
)

func TestUser_Routes_Registration(t *testing.T) {
	// 1. Инициализируем минимальное окружение (mux + huma)
	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("User Test API", "1.0.0")
	api := humago.New(mux, humaConfig)

	// Нам не нужен реальный сервис, так как Huma должна отсечь пустой запрос
	// на этапе валидации схемы (400 Bad Request).
	u := &User{}

	// 2. Регистрируем публичные роуты пользователя
	u.RegistryRoutes(api)

	// Список ручек из ТЗ для этого слайса
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/user/register"},
		{http.MethodPost, "/api/user/login"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			// Создаем пустой запрос без Body
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rr := httptest.NewRecorder()

			// Направляем запрос в мультиплексор
			mux.ServeHTTP(rr, req)

			// Проверка 1: Ручка найдена (не 404)
			assert.NotEqual(t, http.StatusNotFound, rr.Code, "Маршрут %s %s не зарегистрирован", rt.method, rt.path)

			// Проверка 2: Ручка публичная и требует валидный Body (400 Bad Request)
			// Если бы ручка требовала Auth, мы бы получили 401.
			// Здесь мы ожидаем 400, так как Huma видит пустой Body для обязательных полей Login/Password.
			assert.Equal(t, http.StatusBadRequest, rr.Code, "Маршрут %s %s должен быть публичным и валидировать Body", rt.method, rt.path)
		})
	}
}
