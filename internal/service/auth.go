package service

import (
	"Loyalty/internal/models"
	"Loyalty/internal/repository"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Auth struct {
	Repository
}

func NewAuth(repos *repository.Repository) *Auth {
	return &Auth{Repository: repos}
}

func (a *Auth) CreateUser(user *models.User, accountNumber uint64) error {
	//hashing the password
	user.Password = a.HashingPassword(user.Password)

	//try saving user in DB
	err := a.Repository.SaveUser(user, accountNumber)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (a *Auth) GenerateTokenPair(login string) (string, string, error) {
	//create access token
	expTime := viper.GetInt("auth.accessTokenExp")
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        login,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(expTime)).Unix(),
		Subject:   "access",
	})
	token, err := claims.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		logrus.Error(err)
		return "", "", err
	}
	//create refresh token
	rtClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        login,
		ExpiresAt: time.Now().Add(time.Hour * 10000).Unix(),
		Subject:   "refresh",
	})
	rToken, err := rtClaims.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		logrus.Error(err)
		return "", "", err
	}
	return token, rToken, nil
}

func (a *Auth) HashingPassword(password string) string {
	h := sha1.New()
	h.Write([]byte(password))
	hash := h.Sum([]byte(os.Getenv("SALT")))
	return fmt.Sprintf("%x", hash)
}

func (a *Auth) ValidateToken(bearertoken string, tokenType string) (string, error) {
	//validate token
	token, err := jwt.ParseWithClaims(bearertoken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	//read claims
	claims := token.Claims.(*jwt.StandardClaims)
	//check token type
	if claims.Subject != tokenType {
		return "", errors.New("error: not found valid token")
	}

	return claims.Id, nil
}
