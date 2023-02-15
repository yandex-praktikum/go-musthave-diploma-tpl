package router

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserBalance(c echo.Context) error {
	get := c.Get("user")
	points, err := s.db.ReadUserPoints(get.(string))
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	allOrderJSON, err := json.Marshal(points)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(allOrderJSON)
	return nil
}
