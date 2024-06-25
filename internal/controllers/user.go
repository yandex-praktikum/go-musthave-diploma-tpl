package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/application"
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func UserRegister() echo.HandlerFunc {
	return func(c echo.Context) error {
		var userRegisterRequest models.UserRegisterRequest
		err := c.Bind(&userRegisterRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusBadRequest, nil)
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err = validate.Struct(userRegisterRequest)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ExtractErrors(err))
		}

		existUser, err := application.App.UserRepository.FindBy(models.UserSearchFilter{Login: userRegisterRequest.Login})
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}
		if existUser != nil {
			c.Logger().Error("login already exist")
			return c.JSON(http.StatusConflict, "login already exist")
		}

		user, err := application.App.UserRepository.Create(userRegisterRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}

		err = auth.GenerateTokensAndSetCookies(c, user)
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "internal gophermart error")
		}

		return c.JSON(http.StatusOK, user)
	}
}

func UserLogin() echo.HandlerFunc {
	return func(c echo.Context) error {
		var userLoginRequest models.UserLoginRequest
		err := c.Bind(&userLoginRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusBadRequest, nil)
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err = validate.Struct(userLoginRequest)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ExtractErrors(err))
		}

		userRepository := application.App.UserRepository
		existUser, err := userRepository.FindBy(models.UserSearchFilter{Login: userLoginRequest.Login})
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}
		if existUser == nil {
			c.Logger().Error("user not exist")
			return c.JSON(http.StatusUnauthorized, "user not exist")
		}

		if bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(userLoginRequest.Password)) != nil {
			c.Logger().Error("invalid password")
			return c.JSON(http.StatusUnauthorized, "invalid password")
		}

		err = auth.GenerateTokensAndSetCookies(c, &models.UserInfoResponse{
			ID:         existUser.ID,
			LastName:   existUser.LastName,
			FirstName:  existUser.FirstName,
			MiddleName: existUser.MiddleName,
			Login:      existUser.Login,
			Email:      existUser.Email,
		})
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "internal gophermart error")
		}

		return c.JSON(http.StatusOK, models.MapUserToUserLoginResponse(existUser))
	}
}
