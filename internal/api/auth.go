package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kindenko/gophermart/internal/models"
	"github.com/kindenko/gophermart/internal/utils"

	"github.com/kindenko/gophermart/internal/auth"
	"github.com/labstack/echo/v4"
)

func (s *Server) UserAuthorization(c echo.Context) error {

	var userReq models.User
	var userDB models.User

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %s", err)
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	err = json.Unmarshal(b, &userReq)
	if err != nil {
		log.Printf("Failed to parse JSON: %s", err)
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	row := s.DB.First(&userDB, models.User{Login: userReq.Login})
	if row.Error != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid Login or Password")
	}

	verifyPass := utils.VerifyPassword(userDB.Password, userReq.Password)
	if verifyPass != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid Login or Password")
	}

	token, err := auth.GenerateToken(userDB)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "inside error")
	}

	expiresTime := time.Now().Add(3 * time.Hour)
	generateCookie(userReq.Login, token, expiresTime, c)

	return c.String(http.StatusOK, "Bad Request")

}
