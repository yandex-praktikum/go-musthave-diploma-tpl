package router

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserWithdrawals(c echo.Context) error {
	get := c.Get("user")
	allOrder, err := s.db.ReadAllOrderWithdrawnUser(get.(string))
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

	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allOrderJSON)
	return nil
}
