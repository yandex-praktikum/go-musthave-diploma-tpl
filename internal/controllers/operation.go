package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/application"
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// CreateWithdraw Cписание баллов с накопительного счёта в счёт оплаты нового заказа
func CreateWithdraw() echo.HandlerFunc {
	return func(c echo.Context) error {
		var createWithdrawRequest models.CreateWithdrawRequest
		err := c.Bind(&createWithdrawRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, nil)
		}

		currentUserID := auth.GetUserID(c)
		if currentUserID == 0 {
			c.Logger().Error("Unauthorized user create withdraw")
			return c.JSON(http.StatusUnauthorized, nil)
		}

		bonusAccount, err := application.App.AccountRepository.FindByUserID(currentUserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			c.Logger().Error("Can't find bonus account", err)
			return c.JSON(http.StatusInternalServerError, nil)
		}
		if bonusAccount.Sum < createWithdrawRequest.Sum {
			c.Logger().Error("Balance error")
			return c.JSON(http.StatusPaymentRequired, nil)
		}

		order, err := application.App.OrderRepository.FindByNumber(createWithdrawRequest.Order)
		if err != nil || (order != nil && order.UserID != currentUserID) {
			c.Logger().Error("Can't find order", createWithdrawRequest.Order, currentUserID, err)
			return c.JSON(http.StatusUnprocessableEntity, nil)
		}

		err = application.App.OperationRepository.CreateWithdrawn(bonusAccount.ID, createWithdrawRequest.Order, createWithdrawRequest.Sum)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		return c.JSON(http.StatusOK, nil)
	}
}

// GetWithdrawals Получение информации о выводе средств
func GetWithdrawals() echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUserID := auth.GetUserID(c)
		if currentUserID == 0 {
			c.Logger().Error("Unauthorized user create withdraw")
			return c.JSON(http.StatusUnauthorized, nil)
		}

		bonusAccount, err := application.App.AccountRepository.FindByUserID(currentUserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		operations, err := application.App.OperationRepository.GetWithdrawalsByAccountID(bonusAccount.ID)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if len(operations) == 0 {
			return c.JSON(http.StatusNoContent, nil)
		}

		return c.JSON(http.StatusOK, operations)
	}
}
