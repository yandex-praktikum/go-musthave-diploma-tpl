package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserBalance(c echo.Context) error {
	get := c.Get("user")
	points, err := s.DB.ReadUserPoints(get.(string))
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	allPointsJSON, err := json.Marshal(points)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	fmt.Println("====getAPIUserBalance===", allPointsJSON)
	c.Response().Header().Set("content-type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allPointsJSON)
	return nil
}
