package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo"

	"GopherMart/internal/events"
)

func (s *serverMart) postAPIUserLogin(c echo.Context) error {
	var userLog registration
	defer c.Request().Body.Close()
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}
	if err = json.Unmarshal(body, &userLog); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tokenJWT, err := s.DB.LoginUser(ctx, userLog.Login, userLog.Password)
	if (tokenJWT == "") && (err == nil) {
		c.Response().WriteHeader(http.StatusUnauthorized)
		return nil
	}

	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	c.Response().Header().Set(events.Authorization, events.Bearer+" "+tokenJWT)
	c.Response().WriteHeader(http.StatusOK)
	return nil
}
