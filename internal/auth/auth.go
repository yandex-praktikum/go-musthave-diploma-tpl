package auth

import (
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	userTokenCookieName    = "user"
	accessTokenCookieName  = "access-token"
	refreshTokenCookieName = "refresh-token"
)

func GetJWTSecret() string {
	jwtSecretKey, exists := os.LookupEnv("JWT_SECRET_KEY")

	if !exists {
		panic("no JWT_SECRET_KEY in .env")
	}
	return jwtSecretKey
}

func GetAccessTokenCookieName() string {
	return accessTokenCookieName
}

func GetRefreshTokenCookieName() string {
	return refreshTokenCookieName
}

type Claims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

func BeforeFunc(c echo.Context) {
	accessTokenCookie, err := c.Cookie(GetAccessTokenCookieName())
	if err == nil {
		accessUserID, _ := getUserIDByCookie(c, accessTokenCookie)
		if accessUserID != nil {
			return
		}
	}

	refreshTokenCookie, err := c.Cookie(GetRefreshTokenCookieName())
	if err != nil {
		return
	}

	if accessTokenCookie == nil && refreshTokenCookie != nil {
		userID, err := getUserIDByCookie(c, refreshTokenCookie)
		if err != nil {
			return
		}

		user := getUserByID(c, *userID)

		err = GenerateTokensAndSetCookies(c, user)
		if err != nil {
			return
		}
	}
}

func GenerateTokensAndSetCookies(c echo.Context, user *models.UserInfoResponse) error {
	accessToken, accessTokenString, exp, err := generateAccessToken(user)
	if err != nil {
		return err
	}

	_, refreshTokenString, refreshExp, err := generateRefreshToken(user)
	if err != nil {
		return err
	}

	setTokenCookie(c, accessTokenCookieName, accessTokenString, exp)
	setTokenCookie(c, refreshTokenCookieName, refreshTokenString, refreshExp)
	c.Set("user", accessToken)
	setUserCookie(c, user, exp)

	return nil
}

func GetUserID(c echo.Context) uint {
	if c.Get("user") == nil {
		return 0
	}
	u := c.Get("user").(*jwt.Token)

	claims := u.Claims.(*Claims)

	return claims.ID
}

func JWTErrorChecker(c echo.Context, err error) error {
	if err != nil {
		zap.L().Error(
			"JWTErrorChecker",
			zap.Error(err),
		)
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
}

func generateAccessToken(user *models.UserInfoResponse) (*jwt.Token, string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	return generateToken(user, expirationTime, []byte(GetJWTSecret()))
}

func generateRefreshToken(user *models.UserInfoResponse) (*jwt.Token, string, time.Time, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	return generateToken(user, expirationTime, []byte(GetJWTSecret()))
}

func generateToken(user *models.UserInfoResponse, expirationTime time.Time, secret []byte) (*jwt.Token, string, time.Time, error) {
	claims := &Claims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, "", time.Now(), err
	}

	return token, tokenString, expirationTime, nil
}

func setTokenCookie(c echo.Context, name, token string, expiration time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = token
	cookie.Expires = expiration
	cookie.Path = "/"
	cookie.HttpOnly = true

	c.SetCookie(cookie)
}

func setUserCookie(c echo.Context, user *models.UserInfoResponse, expiration time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = userTokenCookieName
	cookie.Value = strconv.Itoa(int(user.ID))
	cookie.Expires = expiration
	cookie.Path = "/"
	c.SetCookie(cookie)
}
