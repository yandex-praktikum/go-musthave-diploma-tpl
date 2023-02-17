package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserWithdrawals(c echo.Context) error {
	get := c.Get("user")
	allOrder, err := s.DB.ReadAllOrderWithdrawnUser(get.(string))
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
	fmt.Println("====getAPIUserWithdrawals==== ", string(allOrderJSON), " ", get.(string))
	c.Response().Header().Set("content-type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allOrderJSON)
	return nil
}
