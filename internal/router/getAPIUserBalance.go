package router

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserBalance(c echo.Context) error {
	get := c.Get("user")
	points, err := s.DB.ReadUserPoints(c.Request().Context(), get.(string))
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	allPointsJSON, err := json.Marshal(points)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	c.Response().Header().Set("content-type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allPointsJSON)
	return nil
}
