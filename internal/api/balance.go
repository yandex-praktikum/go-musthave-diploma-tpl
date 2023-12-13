package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/kindenko/gophermart/internal/auth"
	"github.com/kindenko/gophermart/internal/models"
)

func (s *Server) GetUserBalace(c echo.Context) error {

	var ordersAccrual models.Balance
	var ordersWithdrawn models.Balance
	var balance models.Balance

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

	rowAccrual := s.DB.Table("orders").Select("user_login, sum(accrual) as Current").Where("user_login = ? and operation_type = ?", userLogin, "Accrual").Group("user_login").Find(&ordersAccrual)
	if rowAccrual.Error != nil {
		log.Printf("Failed queries in DB, %s", rowAccrual.Error)
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	rowWithdrawn := s.DB.Table("orders").Select("user_login, sum(accrual) as Withdrawn").Where("user_login = ? and operation_type = ?", userLogin, "Withdrawn").Group("user_login").Find(&ordersWithdrawn)
	if rowWithdrawn.Error != nil {
		log.Printf("Failed queries in DB, %s", rowAccrual.Error)
		return c.JSON(http.StatusInternalServerError, "Internal error")
	}

	balance.Current = ordersAccrual.Current - ordersWithdrawn.Withdrawn
	balance.Withdrawn = ordersWithdrawn.Withdrawn

	return c.JSON(http.StatusOK, balance)
}
