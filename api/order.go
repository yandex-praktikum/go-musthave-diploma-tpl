package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/A-Kuklin/gophermart/api/types"
	"github.com/A-Kuklin/gophermart/internal/auth"
	"github.com/A-Kuklin/gophermart/internal/usecases"
	"github.com/gin-gonic/gin"
)

var (
	ErrEmptyOrderBody = errors.New("CreateOrder body error")
)

func (a *API) CreateOrder(gCtx *gin.Context) {
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

	body, err := io.ReadAll(gCtx.Request.Body)
	if err != nil {
		a.logger.WithError(err).Info("CreateOrder body error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if len(body) == 0 {
		a.logger.Info("CreateOrder empty body")
		gCtx.AbortWithError(http.StatusBadRequest, ErrEmptyOrderBody)
		return
	}

	strOrder := string(body)

	order, err := a.Usecases.Order.Create(gCtx, userID, strOrder)
	switch {
	case errors.Is(err, usecases.ErrLuhnCheck):
		resp := new(types.ResponseMeta).FromMessage("order_id", "Invalid order format")
		gCtx.AbortWithStatusJSON(http.StatusUnprocessableEntity, resp)
		return
	case errors.Is(err, usecases.ErrCreateExistingOrder):
		resp := new(types.ResponseMeta).FromMessage("order_id", "Order has been created already")
		gCtx.AbortWithStatusJSON(http.StatusOK, resp)
		return
	case errors.Is(err, usecases.ErrUniqueOrder):
		resp := new(types.ResponseMeta).FromMessage("order_id", "Order was created by another user")
		gCtx.AbortWithStatusJSON(http.StatusConflict, resp)
		return
	}

	gCtx.JSON(http.StatusAccepted, types.OrderCreateResponse{Order: order})
}

func (a *API) GetOrders(gCtx *gin.Context) {
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

	ordersEntities, err := a.Usecases.Order.GetOrders(gCtx, userID)
	switch {
	case err != nil:
		a.logger.WithError(err).Info("GetOrders postgres error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	case len(ordersEntities) == 0:
		a.logger.Infof("There are no orders for %s", userID)
		resp := new(types.ResponseMeta).FromMessage("user_id", "There are no orders for user")
		gCtx.AbortWithStatusJSON(http.StatusNoContent, resp)
		return
	}

	ordersList := make([]types.OrderResponse, 0, len(ordersEntities))

	for i := range ordersEntities {
		order := types.NewOrderFromDomain(&ordersEntities[i])
		ordersList = append(ordersList, *order)
	}

	jsonData, err := json.Marshal(ordersList)
	if err != nil {
		a.logger.WithError(err).Info("JSON marshaling error")
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.Writer.Write(jsonData)
}
