package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) setOrder(ctx echo.Context) error {

	// Получаем значение заголовка Content-Type
	contentType := ctx.Request().Header.Get("Content-Type")

	// Проверяем, что Content-Type равен "text/plain"
	if contentType != "text/plain" {
		return ctx.JSON(http.StatusBadRequest, "Content-Type не соответсвует ожидаемому: text/plain")
	}
	//body, err := h.getBody(ctx)
	//if err != nil {
	//	return ctx.String(http.StatusInternalServerError, err.Error())
	//}
	//userID, err := h.getUserID(ctx)
	//if err != nil {
	//	h.Short.Logger.Error("ошибка при работе с телом запроса: " + err.Error())
	//	return ctx.String(http.StatusUnauthorized, err.Error())
	//}
	//defer ctx.Request().Body.Close()
	//res := model.PostURLHandlerResponse{}
	//listURL, err := h.Short.ReaderBody(body, userID, model.One)
	//if err != nil {
	//	res.ErrMsg = err.Error()
	//	if strings.HasPrefix(res.ErrMsg, model.ERRCONFLICT) {
	//		return ctx.JSON(http.StatusConflict, res)
	//	}
	//	return ctx.JSON(http.StatusInternalServerError, res)
	//}
	//for _, couple := range listURL {
	//	res.Result = couple.ShortURL
	//}

	//return ctx.JSON(http.StatusCreated, res)
	return nil
}

func (h *Handlers) getOrderList(ctx echo.Context) error {
	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	_, ok := h.Short.UserCH[login]
	if !ok {
		h.Short.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", login)
	}
	list := h.Short.GetOrderList(login)

	return ctx.JSON(http.StatusOK, list)
}
