package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kindenko/gophermart/internal/auth"
	"github.com/kindenko/gophermart/internal/models"
	"github.com/kindenko/gophermart/internal/utils"

	"github.com/labstack/echo/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {

		cookie, err := c.Cookie("Token")
		if err != nil {
			return c.String(http.StatusUnauthorized, "Register before you start working")
		}

		if auth.ValidationToken(string(cookie.Value)) {
			return next(c)
		}
		return c.String(http.StatusUnauthorized, "Register before you start working")
	}

}

func (s *Server) UserAuthentication(c echo.Context) error {
	var userReq models.Users
	var userDB models.Users

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %s", err)
		return c.String(http.StatusBadRequest, "Invalid request type")
	}

	err = json.Unmarshal(b, &userReq)
	if err != nil {
		log.Printf("Failed to parse JSON: %s", err)
		return c.String(http.StatusBadRequest, "Invalid request type")
	}

	if userReq.Login == "" || userReq.Password == "" {
		log.Println("Invalid request type")
		return c.String(http.StatusBadRequest, "Invalid request type")
	}

	row := s.DB.First(&userDB, models.Users{Login: userReq.Login})
	if row.Error != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	if err = utils.VerifyPassword(userDB.Password, userReq.Password); userDB.Login != userReq.Login || err != nil {
		return c.String(http.StatusUnauthorized, "Invalid login/password")
	}

	expiresTime := time.Now().Add(3 * time.Hour)
	token, err := auth.GenerateToken(userDB)

	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	utils.GenerateCookie(token, expiresTime, c)

	return c.String(http.StatusOK, "User is authenticated")

}
