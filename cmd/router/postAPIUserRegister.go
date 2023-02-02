package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type register struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (s *serverMart) postAPIUserRegister(c echo.Context) error {
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

	var pgErr *pgconn.PgError

	hexCookie, err := s.db.RegisterUser(userLog.Login, userLog.Password)
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation: // дубликат
			c.Response().WriteHeader(http.StatusConflict)
			return nil
		default:
			c.Response().WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	cookie := new(http.Cookie)
	cookie.Name = "GopherMart"
	cookie.Value = hexCookie
	cookie.Domain = userLog.Login
	c.SetCookie(cookie)

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
