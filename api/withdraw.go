package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/A-Kuklin/gophermart/api/types"
	"github.com/A-Kuklin/gophermart/internal/auth"
)

func (a *API) Withdraw(gCtx *gin.Context) {
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

	req := types.WithdrawCreateRequest{}

	if err = gCtx.ShouldBindJSON(&req); err != nil {
		a.logger.WithError(err).Info("failed to bind create withdraw request")
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	withdrawArgs, err := req.ToDomain()
	withdrawArgs.UserID = userID

	accrualSum, err := a.Usecases.Order.GetAccruals(gCtx, userID)
	switch {
	case err != nil:
		a.logger.WithError(err).Info("GetAccruals error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	case withdrawArgs.Amount > accrualSum:
		resp := new(types.ResponseMeta).FromMessage("sum", "Insufficient funds")
		gCtx.AbortWithStatusJSON(http.StatusPaymentRequired, resp)
		return
	}

	withdraw, err := a.Usecases.Withdraw.Create(gCtx, withdrawArgs)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
		resp := new(types.ResponseMeta).FromMessage("withdraw", "Order already exists")
		gCtx.AbortWithStatusJSON(http.StatusUnprocessableEntity, resp)
		return
	case err != nil:
		a.logger.WithError(err).Info("Create withdraw error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, withdraw)
}

func (a *API) GetWithdrawals(gCtx *gin.Context) {
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

	withdrawalsEnt, err := a.Usecases.Withdraw.GetWithdrawals(gCtx, userID)
	switch {
	case err != nil:
		a.logger.WithError(err).Info("GetWithdrawals postgres error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	case len(withdrawalsEnt) == 0:
		a.logger.Infof("There are no withdrawals for %s", userID)
		resp := new(types.ResponseMeta).FromMessage("user_id", "There are no withdrawals for user")
		gCtx.AbortWithStatusJSON(http.StatusNoContent, resp)
		return
	}

	withdrawalsList := make([]types.WithdrawResponse, 0, len(withdrawalsEnt))

	for i := range withdrawalsEnt {
		withdraw := types.NewWithdrawFromDomain(&withdrawalsEnt[i])
		withdrawalsList = append(withdrawalsList, *withdraw)
	}

	jsonData, err := json.Marshal(withdrawalsList)
	if err != nil {
		a.logger.WithError(err).Info("JSON marshaling error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.Writer.Write(jsonData)
}
