package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// setOrder - установить нового заказа
func (h *Handlers) setOrder(ctx echo.Context) error {

	// Получаем значение заголовка Content-Type
	contentType := ctx.Request().Header.Get("Content-Type")

	// Проверяем, что Content-Type равен "text/plain"
	if contentType != "text/plain" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Content-Type не соответсвует ожидаемому: text/plain"})
	}
	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	h.Market.Mu.RLock()
	user, ok := h.Market.UserCH[login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusUnauthorized, " пользователь с логином %s не существует", login)
	}
	// Прочитываем тело как plain text
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		h.Market.Lg.Error("setOrder.err - не удалось прочитать тело запроса")
		return echo.NewHTTPError(http.StatusBadRequest, "не удалось прочитать тело")
	}

	order := string(body) // луна и невереный формат
	if !isValidLuhn(order) {
		h.Market.Lg.Error("setOrder.err - невереный формат номера заказа: " + order)
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"message": "ошибка - невереный формат номера заказа"})
	}

	id, _ := strconv.Atoi(order)
	_, ok = user.OrderList[id]
	if ok {
		h.Market.Lg.Info("setOrder.exsist - заказ: " + order + "уже зарегестрирован у этого пользователя")
		return ctx.JSON(http.StatusOK, map[string]string{"message": "номер заказа уже был загружен этим пользователем"})
	}

	// проверка на существование заказа у других
	ok, err = h.Market.Repo.OrderExists(h.Market.Ctx, id)
	if err != nil {
		h.Market.Lg.Error(fmt.Sprintf("setOrder.error - ошибка при поиске заказа на наличие в БД %v", order))
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка - " + err.Error()})
	}
	if ok {
		h.Market.Lg.Error(fmt.Sprintf("setOrder.error - номер заказа %v - уже был загружен другим пользователем ", order))
		return ctx.JSON(http.StatusConflict, map[string]string{"message": "номер заказа уже был загружен другим пользователем"})
	}
	err = h.Market.SetOrder(login, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка - " + err.Error()})
	}

	return ctx.JSON(http.StatusAccepted, map[string]string{"message": " новый номер заказа принят в обработку"})
}

// getOrderList - получить список заказов
func (h *Handlers) getOrderList(ctx echo.Context) error {
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
	list := h.Market.GetOrderList(login)
	if len(list) == 0 {
		return ctx.JSON(http.StatusNoContent, "")
	}
	return ctx.JSON(http.StatusOK, list)
}
