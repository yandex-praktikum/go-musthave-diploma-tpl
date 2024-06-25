package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/application"
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetBalance Получение баланса пользователя
func GetBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp := &models.GetBalanceResponse{}
		currentUserID := auth.GetUserID(c)

		bonusAccount, err := application.App.AccountRepository.FindByUserID(currentUserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		resp.Current = bonusAccount.Sum

		withdrawn, err := application.App.OperationRepository.GetWithdrawnByAccountID(bonusAccount.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if withdrawn != 0 {
			resp.Withdrawn = withdrawn
		}

		return c.JSON(http.StatusOK, resp)
	}
}
