package handler

import (
	"fmt"
	"musthave/internal/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

// withdrawPoints - списание бонусов
func (h *Handlers) withdrawPoints(ctx echo.Context) error {

	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	h.Market.Mu.RLock()
	_, ok := h.Market.UserCH[login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", login)
	}
	var req model.WithdrawReq
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверные данные")
	}
	cb, _, err := h.Market.GetMyBalance(login)
	if err != nil {
		h.Market.Lg.Error("ошибка подсчета баланса: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка подсчета баланса: "+err.Error())
	}
	if cb.Cmp(req.Value) < 0 {
		h.Market.Lg.Error("ошибка списания бонусов - баланс меньше суммы списания.", " Баланс: ", cb)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ошибка списания бонусов - баланс меньше суммы списания. Баланс: %v ", cb))

	}
	err = h.Market.WithdrawnBalance(login, req.Order, req.Value)
	if err != nil {
		h.Market.Lg.Error("ошибка списания бонусов: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "неверные данные")
	}

	return ctx.JSON(http.StatusOK, "")
}

// infoWithdrawals - информация по списаниям
func (h *Handlers) infoWithdrawals(ctx echo.Context) error {
	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	h.Market.Mu.RLock()
	_, ok := h.Market.UserCH[login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", login)
	}
	list, err := h.Market.GetInfoWithdrawnBalance(login)
	if err != nil {
		h.Market.Lg.Error("ошибка получения информации по списаниям: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка получения информации по списаниям: "+err.Error())
	}
	if len(list) == 0 {
		h.Market.Lg.Info("нет ни одного списания")
		return ctx.JSON(http.StatusNoContent, "нет ни одного списания")
	}

	return ctx.JSON(http.StatusOK, list)
}
