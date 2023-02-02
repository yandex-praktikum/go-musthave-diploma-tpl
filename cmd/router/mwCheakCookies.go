package router

import (
	"github.com/labstack/echo"
)

func (s serverMart) CheakCookies(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestCookies, err := c.Request().Cookie("GopherMart")
		if err != nil {
			return nil
		}

		s.db.CheakLogin(requestCookies.Value, requestCookies.Domain)
		return nil
	}
}
