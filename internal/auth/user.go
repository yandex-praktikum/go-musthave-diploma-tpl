package auth

import (
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/application"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetCurrentUserID(c echo.Context) (*uint, error) {
	atc, err := c.Cookie(GetAccessTokenCookieName())
	if err != nil {
		c.Logger().Error(err)
		return nil, err
	}

	if atc == nil {
		return nil, nil
	}

	return getUserIDByCookie(c, atc)
}

func GetCurrentUser(c echo.Context) *models.UserInfoResponse {
	atc, err := c.Cookie(GetAccessTokenCookieName())
	if err != nil {
		c.Logger().Error(err)
		return nil
	}

	if atc == nil {
		return nil
	}

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(atc.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil
	})
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		c.Logger().Error(err)
		return nil
	}

	userID := claims.ID
	if err != nil {
		c.Logger().Error(err)
		return nil
	}

	user, err := application.App.UserRepository.Find(userID)
	if err != nil {
		c.Logger().Error(err)
		return nil
	}

	return user
}

func getUserIDByCookie(c echo.Context, cookie *http.Cookie) (*uint, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil
	})
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		c.Logger().Error(err)
		return nil, err
	}

	return &claims.ID, nil
}

func getUserByID(c echo.Context, id uint) *models.UserInfoResponse {
	user, err := application.App.UserRepository.Find(id)
	if err != nil {
		c.Logger().Error(err)
		return nil
	}

	return user
}
