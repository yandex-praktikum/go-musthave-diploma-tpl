package handler

import (
	"fmt"
	"musthave/internal/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) withdrawPoints(ctx echo.Context) error {

	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	_, ok := h.Short.UserCH[login]
	if !ok {
		h.Short.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", login)
	}
	var req model.WithdrawReq
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверные данные")
	}
	cb, _, err := h.Short.GetMyBalance(login)
	if err != nil {
		h.Short.Lg.Error("ошибка подсчета баланса: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка подсчета баланса: "+err.Error())
	}
	if cb.Cmp(req.Value) < 0 {
		h.Short.Lg.Errorf("ошибка списания бонусов - баланс меньше суммы списания. Баланс: %v ", cb)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ошибка списания бонусов - баланс меньше суммы списания. Баланс: %v ", cb))

	}
	err = h.Short.WithdrawnBalance(login, req.Order, req.Value)
	if err != nil {
		h.Short.Lg.Error("ошибка списания бонусов: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "неверные данные")
	}

	return ctx.JSON(http.StatusOK, "")
}

func (h *Handlers) infoWithdrawals(ctx echo.Context) error {
	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	_, ok := h.Short.UserCH[login]
	if !ok {
		h.Short.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", login)
	}
	list, err := h.Short.GetInfoWithdrawnBalance(login)
	if err != nil {
		h.Short.Lg.Error("ошибка получения информации по списаниям: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка получения информации по списаниям: "+err.Error())
	}
	if len(list) == 0 {
		h.Short.Lg.Info("нет ни одного списания")
		return ctx.JSON(http.StatusNoContent, "нет ни одного списания")
	}

	return ctx.JSON(http.StatusOK, list)
}
