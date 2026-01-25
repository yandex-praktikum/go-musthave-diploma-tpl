package handler

import (
	"musthave/internal/model"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// regitsterUser - регистрация пользователя
func (h *Handlers) regitsterUser(ctx echo.Context) error {
	// Получаем значение заголовка Content-Type
	contentType := ctx.Request().Header.Get("Content-Type")

	// Проверяем, что Content-Type равен "application/json"
	if contentType != "application/json" {
		h.Market.Lg.Error("Content-Type не соответсвует ожидаемому: application/json")
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Content-Type не соответсвует ожидаемому: text/plain"})
	}

	var req model.UserReq
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверные данные")
	}

	if req.Login == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "логин и пароль обязательны")
	}

	h.Market.Mu.RLock()
	_, ok := h.Market.UserCH[req.Login]
	h.Market.Mu.RUnlock()
	if ok {
		h.Market.Lg.Info("пользователь с логином уже существует - " + req.Login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s уже существует", req.Login)
	}

	err := h.auth(ctx, req.Login)
	if err != nil {
		h.Market.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	err = h.Market.RegisterUser(req.Login, req.Password)
	if err != nil {
		h.Market.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return ctx.NoContent(http.StatusOK)
}

// logIn - авторизация пользователя
func (h *Handlers) logIn(ctx echo.Context) error {
	// Получаем значение заголовка Content-Type
	contentType := ctx.Request().Header.Get("Content-Type")

	// Проверяем, что Content-Type равен "application/json"
	if contentType != "application/json" {
		h.Market.Lg.Error("Content-Type не соответсвует ожидаемому: application/json")
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Content-Type не соответсвует ожидаемому: application/json"})
	}
	var req model.UserReq
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"message": "неверные данные"})
	}

	if req.Login == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"message": "логин и пароль обязательны"})
	}
	h.Market.Mu.RLock()
	user, ok := h.Market.UserCH[req.Login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("пользователя с логином не существует - " + req.Login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином не существует "+req.Login)
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(req.Password))
	if err != nil {
		h.Market.Lg.Info("пароль не подходит")
		return echo.NewHTTPError(http.StatusUnauthorized, "неверный пароль")
	}

	err = h.auth(ctx, req.Login)
	if err != nil {
		h.Market.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return nil
	//return ctx.JSON(http.StatusCreated, listURL)
}

// getBalance - получить баланс
func (h *Handlers) getBalance(ctx echo.Context) error {
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
	cb, tw, err := h.Market.GetMyBalance(login)
	if err != nil {
		h.Market.Lg.Error("ошибка подсчета баланса: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка подсчета баланса: "+err.Error())
	}
	info := model.BalanceRes{
		SberThx:   cb,
		Withdrawn: tw,
	}

	return ctx.JSON(http.StatusOK, info)
}
