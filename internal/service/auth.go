package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/botaevg/gophermart/internal/models"
	"github.com/botaevg/gophermart/internal/repositories"
	"github.com/dgrijalva/jwt-go/v4"
	"log"
	"time"
)

type Auth struct {
	storage   repositories.Storage
	secretkey string
}

type Claims struct {
	jwt.StandardClaims
	UserID uint `json:"user_id"`
}

func GenerateHashForPass(password string, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(password))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func NewAuth(storage repositories.Storage, secretkey string) Auth {
	return Auth{
		storage:   storage,
		secretkey: secretkey,
	}
}

func (a Auth) RegisterUser(userAPI models.UserAPI, salt string) (string, error) {
	var User models.User
	User.Username = userAPI.Username
	User.Password = GenerateHashForPass(userAPI.Password, salt)

	ID, err := a.storage.CreateUser(User)
	if err != nil {
		return "", err
	}
	User.ID = ID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: User.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(60 * time.Minute)),
			IssuedAt:  jwt.At(time.Now()),
		},
	})
	tokenSigned, err := token.SignedString([]byte(a.secretkey))
	if err != nil {
		log.Print(err)
		return "", err
	}
	return tokenSigned, err
}

func (a Auth) AuthUser(userAPI models.UserAPI, salt string) (string, error) {
	var User models.User
	User.Username = userAPI.Username
	User.Password = GenerateHashForPass(userAPI.Password, salt)
	ID, err := a.storage.GetUser(User)
	if err != nil {
		return "", err
	}
	User.ID = ID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: User.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(60 * time.Minute)),
			IssuedAt:  jwt.At(time.Now()),
		},
	})
	tokenSigned, err := token.SignedString([]byte(a.secretkey))
	if err != nil {
		log.Print("tokenSigned")
		return "", err
	}
	return tokenSigned, err
}