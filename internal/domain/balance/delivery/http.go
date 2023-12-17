package delivery

import (
	"context"
	"errors"
	"net/http"

	"github.com/benderr/gophermart/internal/domain/balance"
	"github.com/benderr/gophermart/internal/httputils"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/benderr/gophermart/internal/utils"
	"github.com/labstack/echo/v4"
)

type BalanceUsecase interface {
	GetBalanceByUser(ctx context.Context, userid string) (*balance.Balance, error)
	Withdraw(ctx context.Context, userid string, number string, withdraw float64) error
}

type SessionManager interface {
	GetUserId(c echo.Context) (string, error)
}

type balanceHandler struct {
	session SessionManager
	logger  logger.Logger
	BalanceUsecase
}

type WithdrawModel struct {
	Order string   `json:"order" validate:"required"`
	Sum   *float64 `json:"sum" validate:"required"`
}

func NewBalanceHandlers(group *echo.Group, bu BalanceUsecase, session SessionManager, l logger.Logger) {
	h := &balanceHandler{
		BalanceUsecase: bu,
		session:        session,
		logger:         l,
	}

	g := group.Group("/api/user")

	g.GET("/balance", h.GetBalanceHandler)
	g.POST("/balance/withdraw", h.WithdrawHandler)
}

func (o *balanceHandler) GetBalanceHandler(c echo.Context) error {
	userid, err := o.session.GetUserId(c)
	if err != nil {
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	bal, err := o.GetBalanceByUser(c.Request().Context(), userid)

	if err != nil {
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	return c.JSON(http.StatusOK, bal)
}

func (b *balanceHandler) WithdrawHandler(c echo.Context) error {

	var w WithdrawModel

	if err := c.Bind(&w); err != nil {
		return c.JSON(http.StatusBadRequest, httputils.Error("invalid request payload"))
	}

	if err := c.Validate(w); err != nil {
		return c.JSON(http.StatusBadRequest, httputils.Error("invalid request payload"))
	}

	err := utils.ValidateOrder(w.Order)

	if err != nil {
		if errors.Is(err, utils.ErrInvalidNumber) {
			return c.JSON(http.StatusUnprocessableEntity, httputils.Error("invalid order number"))
		}
		b.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	userid, err := b.session.GetUserId(c)
	if err != nil {
		b.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	err = b.Withdraw(c.Request().Context(), userid, w.Order, *w.Sum)

	b.logger.Infoln("MY BALANCE", err)

	if err != nil {
		if errors.Is(err, balance.ErrInsufficientFunds) {
			return c.JSON(http.StatusPaymentRequired, httputils.Error("insufficient funds"))
		}

		b.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	return c.JSON(http.StatusOK, httputils.Ok())
}
