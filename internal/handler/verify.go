package handler

import (
	"net/http"
	"strconv"
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
			return []byte(h.secret), nil
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
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка генерации токена - "+err.Error())
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
	return token.SignedString([]byte(h.secret))
}
func isValidLuhn(number string) bool {
	sum := 0
	alternate := false

	// Проходим по строке справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // Если символ не цифра, номер некорректен
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit / 10) + (digit % 10)
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func (h *Handlers) withAuth(handler echo.HandlerFunc) echo.HandlerFunc {
	return h.authMiddleware(handler)
}
