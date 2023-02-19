package router

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"GopherMart/internal/events"
)

func (s *serverMart) mwUserAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		headerAuth := c.Request().Header.Get(events.Authorization)
		if headerAuth == "" {
			c.Response().WriteHeader(http.StatusUnauthorized)
			return nil
		}
		headerParts := strings.Split(headerAuth, " ")
		if len(headerParts) != 2 || headerParts[0] != events.Bearer {
			c.Response().WriteHeader(http.StatusInternalServerError)
			return nil
		}

		if len(headerParts[1]) == 0 {
			c.Response().WriteHeader(http.StatusInternalServerError)
			return nil
		}

		claims, err := events.DecodeJWT(headerParts[1])
		if err != nil {
			c.Response().WriteHeader(http.StatusInternalServerError)
			return nil
		}

		c.Set("user", claims.Login)
		return next(c)
	}
}
