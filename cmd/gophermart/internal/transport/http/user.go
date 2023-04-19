package http

import (
	"errors"
	"net/http"

	"github.com/RedWood011/cmd/gophermart/internal/apperrors"
	"github.com/RedWood011/cmd/gophermart/internal/entity"
	"github.com/RedWood011/cmd/gophermart/internal/transport/http/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserRegRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserBalanceResponse struct {
	Balance float32 `json:"current"`
	Spent   float32 `json:"withdrawn"`
}

func (c *Controller) UserRegistration() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var userRegRequest UserRegRequest
		if err := ctx.BodyParser(&userRegRequest); err != nil {
			ctx.Status(http.StatusBadRequest)
			return ctx.JSON(ErrorResponse(err))
		}
		user := entity.User{
			ID:       utils.UUIDv4(),
			Login:    userRegRequest.Login,
			Password: userRegRequest.Password,
		}
		if !user.IsValidLogin() || !user.IsValidPassword() {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(apperrors.ErrAuth))
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(err))
		}

		user.Password = string(hash)
		err = c.service.CreateUser(ctx.Context(), user)
		if err != nil {
			switch {
			case errors.Is(err, apperrors.ErrUserExists):
				ctx.Status(http.StatusConflict)
			default:
				ctx.Status(http.StatusInternalServerError)
			}
			return ctx.JSON(ErrorResponse(err))
		}
		token, err := auth.CreateTokenJWT(user, c.cfg.Token)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(err))
		}
		ctx.Set("Authorization", "Bearer "+token)
		ctx.Status(http.StatusOK)
		return nil
	}
}

func (c *Controller) UserAuthorization() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var userRegRequest UserRegRequest
		if err := ctx.BodyParser(&userRegRequest); err != nil {
			ctx.Status(http.StatusBadRequest)
			return ctx.JSON(ErrorResponse(err))
		}
		user := entity.User{
			Login:    userRegRequest.Login,
			Password: userRegRequest.Password,
		}
		if !user.IsValidLogin() || !user.IsValidPassword() {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(apperrors.ErrAuth))
		}
		err := c.service.IdentificationUser(ctx.Context(), user)

		if err != nil {
			switch {
			case errors.Is(err, apperrors.ErrAuth):
				ctx.Status(http.StatusUnauthorized)
				return ctx.JSON(ErrorResponse(apperrors.ErrAuth))
			default:
				ctx.Status(http.StatusInternalServerError)
				return ctx.JSON(ErrorResponse(err))
			}
		}
		token, err := auth.CreateTokenJWT(user, c.cfg.Token)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(err))
		}
		ctx.Set("Authorization", "Bearer "+token)
		ctx.Status(http.StatusOK)
		return nil
	}
}

func (c *Controller) GetBalance() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		balance, err := c.service.GetUserBalance(ctx.Context(), ctx.Get("userID"))
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return ctx.JSON(ErrorResponse(err))
		}
		return ctx.JSON(UserBalanceResponse{Balance: balance.Balance, Spent: balance.Spent})
	}
}
