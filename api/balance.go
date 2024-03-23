package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/A-Kuklin/gophermart/api/types"
	"github.com/A-Kuklin/gophermart/internal/auth"
)

func (a *API) GetBalance(gCtx *gin.Context) {
	userID, err := auth.Auth.GetTokenUserID(gCtx)
	switch {
	case errors.Is(err, auth.ErrGetClaims):
		a.logger.WithError(err).Info("Parse user_id uuid from token error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	case err != nil:
		a.logger.WithError(err).Info("GetTokenUserID error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	accrualCents, err := a.Usecases.Order.GetAccruals(gCtx, userID)
	if err != nil {
		a.logger.WithError(err).Info("GetAccruals error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	withdrawalsCents, err := a.Usecases.Withdraw.GetSumWithdrawals(gCtx, userID)
	if err != nil {
		a.logger.WithError(err).Info("GetWithdrawals error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	balance := float64(accrualCents-withdrawalsCents) / 100
	withdrawals := float64(withdrawalsCents) / 100

	gCtx.JSON(http.StatusOK, types.BalanceResponse{Current: balance, Withdrawals: withdrawals})
}
