package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserOrders(c echo.Context) error {
	fmt.Println("===> getAPIUserOrders")
	get := c.Get("user")
	allOrder, err := s.DB.ReadAllOrderAccrualUser(get.(string))
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	fmt.Println("===============test0 ", allOrder)
	if len(allOrder) == 0 {
		fmt.Println("===============test1")
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}

	allOrderJSON, err := json.Marshal(allOrder)
	if err != nil {
		fmt.Println("===============test2")
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	fmt.Println("===============test3 ", string(allOrderJSON))
	c.Response().Header().Set("content-type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allOrderJSON)
	return nil
}
