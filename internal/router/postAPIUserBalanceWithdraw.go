package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"GopherMart/internal/errorsgm"
	"GopherMart/internal/events"
)

type orderWithdrawals struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (s *serverMart) postAPIUserBalanceWithdraw(c echo.Context) error {

	defer c.Request().Body.Close()
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	var bodyOrder orderWithdrawals
	err = json.Unmarshal(body, &bodyOrder)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	atoi, err := strconv.Atoi(bodyOrder.Order)
	if err != nil {
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return nil
	}
	if !events.Valid(atoi) {
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return nil
	}

	get := c.Get("user")
	err = s.DB.WithdrawnUserPoints(c.Request().Context(), get.(string), bodyOrder.Order, bodyOrder.Sum)

	if err != nil {
		if errors.Is(err, errorsgm.ErrDontHavePoints) {
			c.Response().WriteHeader(http.StatusPaymentRequired)
			return nil
		}
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	c.Response().WriteHeader(http.StatusOK)
	return nil
}
