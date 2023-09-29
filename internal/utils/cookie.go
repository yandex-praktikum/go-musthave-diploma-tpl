package utils

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func GenerateCookie(token string, expr time.Time, c echo.Context) {

	cookie := new(http.Cookie)
	cookie.Name = "Token"
	cookie.Value = token
	cookie.Expires = expr
	cookie.Path = "/"
	c.SetCookie(cookie)
}
