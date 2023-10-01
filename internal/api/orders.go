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

	var orderDB models.Orders

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

	row := s.DB.First(&orderDB, models.Orders{Number: string(b)})
	if row.RowsAffected == 0 {
		newOrder := models.Orders{Number: string(b), UserLogin: userLogin, UploadedAt: time.Now(), Accrual: 0, Status: "New", OperationType: "Accrual"}
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
	var orders []models.Orders

	cookie, err := c.Cookie("Token")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}
	userLogin, err := auth.GetUserLoginFromToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	s.DB.Order("uploaded_at").Where("user_login = ? and operation_type = ?", userLogin, "Accrual").Find(&orders)

	if len(orders) == 0 {
		return c.JSON(http.StatusNoContent, "No data to answer")
	}

	return c.JSON(http.StatusOK, orders)
}
