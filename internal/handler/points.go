package handler

import (
	"fmt"
	"musthave/internal/model"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

// withdrawPoints - списание бонусов
func (h *Handlers) withdrawPoints(ctx echo.Context) error {

	login := ctx.Get("user_login").(string)
	if login == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "не смогли вычитать логин из контекста запроса"})
	}
	h.Market.Mu.RLock()
	user, ok := h.Market.UserCH[login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return ctx.JSON(http.StatusConflict, map[string]string{"message": " пользователь с логином не существует " + login})
	}
	var req model.WithdrawReq
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "неверные данные"})
	}
	valueStr := req.Value.String()
	valueDec, err := decimal.NewFromString(valueStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "неверный формат суммы"})
	}
	cb, _, err := h.Market.GetMyBalance(login)
	if err != nil {
		h.Market.Lg.Error("ошибка подсчета баланса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка подсчета баланса: " + err.Error()})
	}
	if cb.Cmp(valueDec) < 0 {
		h.Market.Lg.Error("ошибка списания бонусов - баланс меньше суммы списания.", " Баланс: ", cb)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("ошибка списания бонусов - баланс меньше суммы списания. Баланс: %v ", cb)})

	}
	//
	id, _ := strconv.Atoi(req.Order)
	_, ok = user.OrderList[id]
	if ok {
		h.Market.Lg.Info("setOrder.exsist - заказ: " + req.Order + "уже зарегестрирован у этого пользователя")
		return ctx.JSON(http.StatusOK, map[string]string{"message": "номер заказа уже был загружен этим пользователем"})
	}

	// проверка на существование заказа у других
	ok, err = h.Market.Repo.OrderExists(h.Market.Ctx, id)
	if err != nil {
		h.Market.Lg.Error(fmt.Sprintf("setOrder.error - ошибка при поиске заказа на наличие в БД %v", req.Order))
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка - " + err.Error()})
	}
	if ok {
		h.Market.Lg.Error(fmt.Sprintf("setOrder.error - номер заказа %v - уже был загружен другим пользователем ", req.Order))
		return ctx.JSON(http.StatusConflict, map[string]string{"message": "номер заказа уже был загружен другим пользователем"})
	}
	err = h.Market.SetOrder(login, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка - " + err.Error()})
	}
	//
	err = h.Market.WithdrawnBalance(login, req.Order, valueDec)
	if err != nil {
		h.Market.Lg.Error("ошибка списания бонусов: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "неверные данные"})
	}

	return ctx.JSON(http.StatusOK, "")
}

// infoWithdrawals - информация по списаниям
func (h *Handlers) infoWithdrawals(ctx echo.Context) error {
	login := ctx.Get("user_login").(string)
	if login == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "не смогли вычитать логин из контекста запроса"})
	}
	h.Market.Mu.RLock()
	_, ok := h.Market.UserCH[login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return ctx.JSON(http.StatusConflict, map[string]string{"message": " пользователь с логином не существует" + login})
	}
	list, err := h.Market.GetInfoWithdrawnBalance(login)
	if err != nil {
		h.Market.Lg.Error("ошибка получения информации по списаниям: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка получения информации по списаниям: " + err.Error()})
	}
	if len(list) == 0 {
		h.Market.Lg.Info("нет ни одного списания")
		return ctx.JSON(http.StatusNoContent, map[string]string{"message": "нет ни одного списания"})
	}

	return ctx.JSON(http.StatusOK, list)
}
