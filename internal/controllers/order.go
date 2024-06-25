package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/application"
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/labstack/echo/v4"
	"github.com/theplant/luhn"
	"io"
	"net/http"
	"strconv"
)

func CreateOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		orderNumberString := string(body)
		orderNumber, err := strconv.Atoi(orderNumberString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if !luhn.Valid(orderNumber) {
			return c.JSON(http.StatusUnprocessableEntity, nil)
		}

		existOrder, err := application.App.OrderRepository.FindByNumber(orderNumberString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		currentUserID := auth.GetUserID(c)
		if existOrder != nil {
			if existOrder.UserID == currentUserID {
				return c.JSON(http.StatusOK, nil)
			} else {
				return c.JSON(http.StatusConflict, nil)
			}
		}

		order, err := application.App.OrderRepository.Create(orderNumberString, currentUserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		application.App.AccrualService.SendOrderToQueue(*order)

		return c.JSON(http.StatusAccepted, nil)
	}
}

// GetOrders Получение списка загруженных номеров заказов
func GetOrders() echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUserID := auth.GetUserID(c)

		orders, err := application.App.OrderRepository.GetOrdersByUserID(currentUserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if len(orders) == 0 {
			return c.JSON(http.StatusNoContent, nil)
		}

		return c.JSON(http.StatusOK, orders)
	}
}
