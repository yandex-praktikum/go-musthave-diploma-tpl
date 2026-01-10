package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tartushkin/go-musthave-diploma-tpl.git/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handlers) regitsterUser(ctx echo.Context) error {
	// Получаем значение заголовка Content-Type
	contentType := ctx.Request().Header.Get("Content-Type")

	// Проверяем, что Content-Type равен "application/json"
	if contentType != "application/json" {
		h.Short.Lg.Error("Content-Type не соответсвует ожидаемому: application/json")
		return ctx.JSON(http.StatusBadRequest, "Content-Type не соответсвует ожидаемому: application/json")
	}

	var req model.UserReq
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверные данные")
	}

	if req.Login == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "логин и пароль обязательны")
	}

	_, ok := h.Short.UserCH[req.Login]
	if ok {
		h.Short.Lg.Info("пользователь с логином уже существует - " + req.Login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s уже существует", req.Login)
	}

	err := h.Short.RegisterUser(req.Login, req.Password)
	if err != nil {
		h.Short.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	err = h.auth(ctx, req.Login)
	if err != nil {
		h.Short.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *Handlers) logIn(ctx echo.Context) error {
	// Получаем значение заголовка Content-Type
	contentType := ctx.Request().Header.Get("Content-Type")

	// Проверяем, что Content-Type равен "application/json"
	if contentType != "application/json" {
		h.Short.Lg.Error("Content-Type не соответсвует ожидаемому: application/json")
		return ctx.JSON(http.StatusBadRequest, "Content-Type не соответсвует ожидаемому: application/json")
	}
	var req model.UserReq
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверные данные")
	}

	if req.Login == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "логин и пароль обязательны")
	}
	user, ok := h.Short.UserCH[req.Login]
	if !ok {
		h.Short.Lg.Info("пользователя с логином не существует - " + req.Login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", req.Login)
	}
	passHash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if user.PassHash != string(passHash) {
		h.Short.Lg.Info("пароль не подходит")
		return echo.NewHTTPError(http.StatusUnauthorized, " пароль не подходит")
	}
	err := h.auth(ctx, req.Login)
	if err != nil {
		h.Short.Lg.Error("Ошибка при работе с телом запроса: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	return nil
	//return ctx.JSON(http.StatusCreated, listURL)
}

func (h *Handlers) getBalance(ctx echo.Context) error {
	login := ctx.Get("user_login").(string)
	if login == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "не смогли вычитать логин из контекста запроса")
	}
	_, ok := h.Short.UserCH[login]
	if !ok {
		h.Short.Lg.Info("не нашли пользоваетля с данным логином в кеше - " + login)
		return echo.NewHTTPError(http.StatusConflict, " пользователь с логином %s не существует", login)
	}
	cb, tw, err := h.Short.GetMyBalance(login)
	if err != nil {
		h.Short.Lg.Error("ошибка подсчета баланса: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка подсчета баланса: "+err.Error())
	}
	info := model.BalanceRes{
		SberThx:   cb,
		Withdrawn: tw,
	}

	return ctx.JSON(http.StatusOK, info)
}
