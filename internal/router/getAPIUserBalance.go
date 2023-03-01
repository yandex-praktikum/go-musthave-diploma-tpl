package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

func (s *serverMart) getAPIUserBalance(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	get := c.Get("user")
	points, err := s.DB.ReadUserPoints(ctx, get.(string))
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
