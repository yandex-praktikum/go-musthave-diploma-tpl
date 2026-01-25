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
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "неверные данные"})
	}

	if req.Login == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "логин и пароль обязательны"})
	}

	h.Market.Mu.RLock()
	_, ok := h.Market.UserCH[req.Login]
	h.Market.Mu.RUnlock()
	if ok {
		h.Market.Lg.Info("пользователь с логином уже существует - " + req.Login)
		return ctx.JSON(http.StatusConflict, map[string]string{"message": " пользователь с логином уже существует " + req.Login})
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
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "неверные данные"})
	}

	if req.Login == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "логин и пароль обязательны"})
	}
	h.Market.Mu.RLock()
	user, ok := h.Market.UserCH[req.Login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("пользователя с логином не существует - " + req.Login)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"message": " пользователя с логином не существует " + req.Login})
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(req.Password))
	if err != nil {
		h.Market.Lg.Info("пароль не подходит")
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"message": "неверный пароль"})
	}

	err = h.auth(ctx, req.Login)
	if err != nil {
		h.Market.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return nil
}

// getBalance - получить баланс
func (h *Handlers) getBalance(ctx echo.Context) error {
	login := ctx.Get("user_login").(string)
	if login == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "не смогли вычитать логин из контекста запроса"})
	}
	h.Market.Mu.RLock()
	_, ok := h.Market.UserCH[login]
	h.Market.Mu.RUnlock()
	if !ok {
		h.Market.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return ctx.JSON(http.StatusConflict, map[string]string{"message": " пользователь с логином не существует: " + login})
	}
	cb, tw, err := h.Market.GetMyBalance(login)
	if err != nil {
		h.Market.Lg.Error("ошибка подсчета баланса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ошибка подсчета баланса: " + err.Error()})
	}
	info := model.BalanceRes{
		SberThx:   cb.InexactFloat64(),
		Withdrawn: tw.InexactFloat64(),
	}

	return ctx.JSON(http.StatusOK, info)
}
