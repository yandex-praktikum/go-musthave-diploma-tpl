package delivery

import (
	"context"
	"net/http"

	"github.com/benderr/gophermart/internal/domain/withdrawal"
	"github.com/benderr/gophermart/internal/httputils"
	"github.com/benderr/gophermart/internal/logger"
	"github.com/labstack/echo/v4"
)

type WidthdrawUsecase interface {
	GetWithdrawsByUser(ctx context.Context, userid string) ([]withdrawal.Withdrawal, error)
}

type SessionManager interface {
	GetUserId(c echo.Context) (string, error)
}

type withdrawHandler struct {
	session SessionManager
	logger  logger.Logger
	WidthdrawUsecase
}

func NewWithdrawHandlers(group *echo.Group, wu WidthdrawUsecase, session SessionManager, logger logger.Logger) {
	h := &withdrawHandler{
		WidthdrawUsecase: wu,
		session:          session,
		logger:           logger,
	}

	g := group.Group("/api/user")

	g.GET("/withdrawals", h.GetWithdrawsHandler)
}

func (w *withdrawHandler) GetWithdrawsHandler(c echo.Context) error {
	userid, err := w.session.GetUserId(c)
	if err != nil {
		w.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	list, err := w.GetWithdrawsByUser(c.Request().Context(), userid)

	if err != nil {
		w.logger.Errorln(err)
		return c.JSON(http.StatusInternalServerError, httputils.Error("internal server error"))
	}

	if len(list) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	return c.JSON(http.StatusOK, list)
}
