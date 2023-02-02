package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) postAPIUserLogin(c echo.Context) error {
	var userLog register
	defer c.Request().Body.Close()
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil //400
	}
	if err = json.Unmarshal(body, &userLog); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil // 400
	}

	hexCookie, err := s.db.LoginUser(userLog.Login, userLog.Password)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "GopherMart"
	cookie.Value = hexCookie
	c.SetCookie(cookie)

	return nil
}
