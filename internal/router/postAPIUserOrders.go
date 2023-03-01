package router

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"GopherMart/internal/errorsgm"
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	get := c.Get("user")
	err = s.DB.WriteOrderAccrual(ctx, bodyOrder, get.(string))
	if err != nil {
		if errors.Is(err, errorsgm.ErrLoadedEarlierThisUser) {
			c.Response().WriteHeader(http.StatusOK)
			return nil
		}
		if errors.Is(err, errorsgm.ErrLoadedEarlierAnotherUser) {
			c.Response().WriteHeader(http.StatusConflict)
			return nil
		}

		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	c.Response().WriteHeader(http.StatusAccepted)
	return nil
}
