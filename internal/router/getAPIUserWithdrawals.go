package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserWithdrawals(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	get := c.Get("user")
	allOrder, err := s.DB.ReadAllOrderWithdrawnUser(ctx, get.(string))
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	if len(allOrder) == 0 {
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}

	allOrderJSON, err := json.Marshal(allOrder)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	c.Response().Header().Set("content-type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allOrderJSON)
	return nil
}
