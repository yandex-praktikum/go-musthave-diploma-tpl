package delivery

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/benderr/gophermart/internal/domain/orders"
	"github.com/benderr/gophermart/internal/httputils"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/benderr/gophermart/internal/utils"
	"github.com/labstack/echo/v4"
)

type OrderUsecase interface {
	Create(ctx context.Context, userid string, number string, status orders.Status) (*orders.Order, error)
	GetOrdersByUser(ctx context.Context, userid string) ([]orders.Order, error)
}

type SessionManager interface {
	GetUserId(c echo.Context) (string, error)
}

type ordersHandler struct {
	session SessionManager
	logger  logger.Logger
	OrderUsecase
}

func NewOrderHandlers(group *echo.Group, ou OrderUsecase, session SessionManager, logger logger.Logger) {
	h := &ordersHandler{
		OrderUsecase: ou,
		session:      session,
		logger:       logger,
	}

	g := group.Group("/api/user")

	g.GET("/orders", h.GetOrdersHandler)
	g.POST("/orders", h.CreateOrderHandler)
}

func (o *ordersHandler) GetOrdersHandler(c echo.Context) error {
	userid, err := o.session.GetUserId(c)
	if err != nil {
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.ErrorWithDetails("internal server error", err))
	}

	list, err := o.GetOrdersByUser(c.Request().Context(), userid)

	if err != nil {
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.ErrorWithDetails("internal server error", err))
	}

	if len(list) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	return c.JSON(http.StatusOK, list)
}

func (o *ordersHandler) CreateOrderHandler(c echo.Context) error {

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.ErrorWithDetails("internal server error", err))
	}

	number := string(body)

	if len(number) == 0 {
		return c.JSON(http.StatusBadRequest, httputils.Error("empty number"))
	}

	err = utils.ValidateOrder(number)

	if err != nil {
		if errors.Is(err, utils.ErrInvalidNumber) {
			return c.JSON(http.StatusUnprocessableEntity, httputils.Error("invalid order number"))
		}
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.ErrorWithDetails("internal server error", err))
	}

	userid, err := o.session.GetUserId(c)
	if err != nil {
		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.ErrorWithDetails("internal server error", err))
	}

	_, err = o.Create(c.Request().Context(), userid, number, orders.NEW)

	if err != nil {
		//принят ранее
		if errors.Is(err, orders.ErrExistForUser) {
			return c.JSON(http.StatusOK, httputils.Ok())
		}

		// уже занят другим
		if errors.Is(err, orders.ErrForeignForUser) {
			return c.JSON(http.StatusConflict, httputils.Error("conflict"))
		}

		o.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.ErrorWithDetails("internal server error", err))
	}

	//новый контракт принят
	return c.JSON(http.StatusAccepted, httputils.Ok())
}
