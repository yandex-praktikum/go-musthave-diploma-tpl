package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/constants"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserID int `json:"user_id"`
}

func (h *Handler) SingUp(c *gin.Context) {
	var input models.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, err)
		return
	}
	input.Login = validatelogin(input.Login)

	_, err := h.storage.Autorisation.CreateUser(input)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	token, err := h.storage.Autorisation.GenerateToken(input.Login, input.Password)
	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Writer.Header().Set("Authorization", token)
	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})

}

func (h *Handler) SingIn(c *gin.Context) {
	var input models.User
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, err)
		return
	}
	input.Login = validatelogin(input.Login)

	token, err := h.storage.Autorisation.GenerateToken(input.Login, input.Password)

	if err != nil {
		newErrorResponse(c, err)
		return
	}

	c.Writer.Header().Set("Authorization", token)

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})

}

type AuthService struct {
	// log  logger.Logger
	repo repository.Autorisation
}

func (a *AuthService) CreateUser(user models.User) (int, error) {
	user.Salt = randStr(20)
	user.Password = generatePasswordHash(user.Password, user.Salt)

	return a.repo.CreateUser(user)
}

func NewAuthStorage(repo repository.Autorisation) *AuthService {
	return &AuthService{repo: repo}
}

func (a *AuthService) GenerateToken(username, password string) (string, error) {

	user, err := a.repo.GetUser(username)
	if err != nil {
		return "", err
	}

	inputpass := generatePasswordHash(password, user.Salt)

	if inputpass != user.Password {
		return "", errors.New("unauthorized")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(constants.TokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})

	return token.SignedString([]byte(constants.SigningKey))
}

func (a *AuthService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(constants.SigningKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserID, nil
}
