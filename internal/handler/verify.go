package handler

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// authMiddleware - проверка на авторизацию пользователя
func (h *Handlers) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := ""

		// Из куки
		if cookie, err := c.Cookie("auth"); err == nil {
			tokenString = cookie.Value
		} else {
			// Или из заголовка
			auth := c.Request().Header.Get("Authorization")
			if len(auth) > 7 && auth[:7] == "Bearer " {
				tokenString = auth[7:]
			}
		}

		if tokenString == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "нет токена")
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			return h.secret, nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "неверный токен")
		}

		c.Set("user_login", claims.Subject)
		return next(c)
	}
}

// auth - авторизация пользователя
func (h *Handlers) auth(c echo.Context, log string) error {
	// Генерируем JWT
	token, err := h.generateJWT(log)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка генерации токена")
	}

	// Вариант A: выдаём в куке (для веба)
	c.SetCookie(&http.Cookie{
		Name:     "auth",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24, // 24 часа
	})
	return nil
}

// generateJWT - генерация JWT
func (h *Handlers) generateJWT(log string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   log,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.secret)
}
