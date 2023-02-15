package router

import (
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"GopherMart/internal/errorsGM"
	"GopherMart/internal/events"
)

func (s *serverMart) postAPIUserOrders(c echo.Context) error {
	defer c.Request().Body.Close()
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	bodyOrder := string(body)
	atoi, err := strconv.Atoi(bodyOrder)
	if err != nil {
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return nil
	}
	if !events.Valid(atoi) {
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return nil
	}

	get := c.Get("user")
	err = s.db.WriteOrderAccrual(bodyOrder, get.(string))
	if err != nil {
		if errors.Is(err, errorsGM.ErrLoadedEarlierThisUser) {
			c.Response().WriteHeader(http.StatusOK)
			return nil
		}
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	c.Response().WriteHeader(http.StatusAccepted)
	return nil
}
