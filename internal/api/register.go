package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kindenko/gophermart/internal/auth"
	"github.com/kindenko/gophermart/internal/models"
	"github.com/kindenko/gophermart/internal/utils"
	"github.com/labstack/echo/v4"
)

func (s *Server) UsTests(c echo.Context) error {

	cookie, err := c.Cookie("Token")
	if err != nil {

	}
	userLogin, err := auth.GetUserLoginFromToken(cookie.Value)
	fmt.Println(userLogin, err)

	return c.String(http.StatusAccepted, "прет сук")
}

func (s *Server) UserRegistrater(c echo.Context) error {

	var userReq models.User
	var userDB models.User

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %s", err)
		return c.String(http.StatusBadRequest, "")
	}

	err = json.Unmarshal(b, &userReq)
	if err != nil {
		log.Printf("Failed to parse JSON: %s", err)
		return c.String(http.StatusBadRequest, "")
	}

	if userReq.Login == "" || userReq.Password == "" {
		log.Println("Incorrect request")
		return c.String(http.StatusBadRequest, "Incorrect request")
	}

	row := s.DB.First(&userDB, models.User{Login: userReq.Login})
	if row.Error == nil && userDB.Login == userReq.Login {
		return c.JSON(http.StatusConflict, "The user has already been registered")
	}

	hashPassword, err := utils.HashPassword(userReq.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	newUser := models.User{Login: userReq.Login, Password: hashPassword, Balance: 0}
	result := s.DB.Create(&newUser)

	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to write to user database")
	}

	expiresTime := time.Now().Add(3 * time.Hour)
	token, err := auth.GenerateToken(userDB)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "inside error")
	}

	utils.GenerateCookie(token, expiresTime, c)

	return c.JSON(http.StatusOK, "User has registered!")

}
