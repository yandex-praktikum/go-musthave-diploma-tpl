package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kindenko/gophermart/internal/auth"
	"github.com/kindenko/gophermart/internal/models"
	"github.com/labstack/echo/v4"
)

func (s *Server) WithdrawnPoints(c echo.Context) error {

	var balance models.Balance
	var withdraw models.Withdraw

	cookie, err := c.Cookie("Token")
	if err != nil {
		log.Printf("Could not get cookies, %s", err)
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	userLogin, err := auth.GetUserLoginFromToken(cookie.Value)
	if err != nil {
		log.Printf("Could get user from cookies, %s", err)
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %s", err)
		return c.String(http.StatusBadRequest, "Invalid request type")
	}

	err = json.Unmarshal(b, &withdraw)
	if err != nil {
		log.Printf("Failed to parse JSON: %s", err)
		return c.String(http.StatusBadRequest, "Invalid request type")
	}

	_, err = strconv.Atoi(withdraw.Order)
	if err != nil {
		return c.String(http.StatusUnprocessableEntity, "Invalid order")
	}

	rowAccrual := s.DB.Table("orders").Select("user_login, sum(accrual) as Current").Where("user_login = ? and operation_type = ?", userLogin, "Accrual").Group("user_login").Find(&balance)
	if rowAccrual.Error != nil {
		log.Printf("Failed queries in DB, %s", rowAccrual.Error)
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	rowWithdrawn := s.DB.Table("orders").Select("user_login, sum(accrual) as Withdrawn").Where("user_login = ? and operation_type = ?", userLogin, "Withdrawn").Group("user_login").Find(&balance)
	if rowWithdrawn.Error != nil {
		log.Printf("Failed queries in DB, %s", rowAccrual.Error)
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	if balance.Current-balance.Withdrawn-withdraw.Sum < 0 {
		return c.JSON(http.StatusPaymentRequired, "Insufficient funds")
	}

	newWithdrawn := models.Orders{Number: withdraw.Order, UserLogin: userLogin, UploadedAt: time.Now(), Accrual: withdraw.Sum, Status: "New", OperationType: "Withdrawn"}
	result := s.DB.Create(&newWithdrawn)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to write order in database")
	}

	return c.JSON(http.StatusOK, "Successful request processing")
}
