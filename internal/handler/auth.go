package handler

import (
	"Loyalty/internal/models"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

type UpdateRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type signInResponse struct {
	Login         string `json:"login"`
	AccountNumber uint64 `json:"account_number"`
	RefreshToken  string `json:"refresh_token"`
}

func (h *Handler) AuthMiddleware(c *gin.Context) {
	//read header
	authHeader := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authHeader) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		c.Abort()
		return
	}
	bearerToken := authHeader[1]
	//validate token
	login, err := h.service.ValidateToken(bearerToken, "access")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		c.Abort()
		return
	}
	h.userLogin = login
	c.Next()
}

//Registration
func (h *Handler) SignIn(c *gin.Context) {
	var user models.User
	//parse request
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//validation request
	if ok, _ := govalidator.ValidateStruct(user); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not allowed request"})
		return
	}
	if (len(user.Password) < 7) || (len(user.Password) > 20) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password length will be from 7 to 15 simbols"})
		return
	}

	accountNumber, err := h.service.CreateLoyaltyAccount(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//save in db
	if err := h.service.Auth.CreateUser(&user, accountNumber); err != nil {
		//login conflict
		if errors.Is(err, errors.Unwrap(err)) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		//internal error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//create tokens
	t, rt, err := h.service.Auth.GenerateTokenPair(user.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	authHeader := fmt.Sprint("Bearer ", t)
	c.Header("Authorization", authHeader)

	var resp signInResponse
	resp.Login = user.Login
	resp.AccountNumber = accountNumber
	resp.RefreshToken = rt

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SignUp(c *gin.Context) {
	var user models.User
	//parse request
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//check user in db
	user.Password = h.service.Auth.HashingPassword(user.Password)
	number, err := h.service.Repository.GetUser(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//create tokens
	t, rt, err := h.service.Auth.GenerateTokenPair(user.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	authHeader := fmt.Sprint("Bearer ", t)
	c.Header("Authorization", authHeader)

	var resp signInResponse
	resp.Login = user.Login
	resp.AccountNumber = number
	resp.RefreshToken = rt

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) TokenRefreshing(c *gin.Context) {
	var request UpdateRequest
	//read refresh token
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not allowed request"})
		return
	}
	//validate token
	login, err := h.service.ValidateToken(request.RefreshToken, "refresh")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not valid refresh token"})
		return
	}
	//if validate is ok -> create new tokens
	t, rt, err := h.service.Auth.GenerateTokenPair(login)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	//sent response
	authHeader := fmt.Sprint("Bearer ", t)
	c.Header("Authorization", authHeader)
	c.JSON(http.StatusOK, gin.H{"refresh token": rt})
}
