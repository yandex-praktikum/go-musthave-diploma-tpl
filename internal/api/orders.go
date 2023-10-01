package api

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kindenko/gophermart/internal/auth"
	"github.com/kindenko/gophermart/internal/models"
	"github.com/labstack/echo/v4"
)

func (s *Server) UploadOrder(c echo.Context) error {

	var orderDB models.Order

	cookie, err := c.Cookie("Token")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}
	userLogin, err := auth.GetUserLoginFromToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %s", err)
		return c.String(http.StatusBadRequest, "")
	}

	_, err = strconv.Atoi(string(b))
	if err != nil {
		return c.String(http.StatusUnprocessableEntity, "Invalid order")
	}

	// if !utils.Valid(order) {
	// 	return c.String(http.StatusUnprocessableEntity, "Invalid number order luhn")
	// }

	row := s.DB.First(&orderDB, models.Order{Number: string(b)})
	if row.RowsAffected == 0 {
		newOrder := models.Order{Number: string(b), UserLogin: userLogin, UploadedAt: time.Now(), Status: "New", Accrual: 0}
		result := s.DB.Create(&newOrder)

		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to write order to database")
		}

		return c.String(http.StatusOK, "Order loaded")
	}

	if orderDB.UserLogin != userLogin {
		return c.JSON(http.StatusConflict, "The order was uploaded by another user")
	}

	return c.JSON(http.StatusOK, "The order number has already been uploaded by this user")
}

func (s *Server) GetOrders(c echo.Context) error {
	var orders []models.Order

	cookie, err := c.Cookie("Token")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}
	userLogin, err := auth.GetUserLoginFromToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	s.DB.Order("uploaded_at").Where("user_login = ?", userLogin).Find(&orders)
	if len(orders) == 0 {
		return c.JSON(http.StatusNoContent, "No data to answer")
	}

	return c.JSON(http.StatusOK, orders)
}
