package auth

import (
	"errors"
	"fmt"
	"github.com/A-Kuklin/gophermart/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	UserID               = "user_id"
	AccessTokenHeaderKey = "Authorization"
)

var (
	ErrParsingToken = errors.New("parse client token err")
	ErrGetClaims    = errors.New("failed to parse claims")
)

var Auth *Guard

var signMethod = jwt.SigningMethodHS256

type Guard struct {
	cfg *config.Config
}

func (g *Guard) GetToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(signMethod, jwt.MapClaims{
		UserID: userID,
	})

	signedTokenString, err := token.SignedString([]byte(g.cfg.SecretKey))
	if err != nil {
		return "", fmt.Errorf("sign token err: %w", err)
	}

	return signedTokenString, nil
}

func Init(cfg *config.Config) {
	Auth = &Guard{cfg: cfg}
}

func (g *Guard) GetTokenUserID(gCtx *gin.Context) (uuid.UUID, error) {
	token := gCtx.Request.Header.Get(AccessTokenHeaderKey)
	tknClaims, err := g.getClaims(token)
	if err != nil {
		return uuid.Nil, err
	}
	strUserID := tknClaims[UserID].(string)
	userID, err := uuid.Parse(strUserID)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (g *Guard) getClaims(strToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(strToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(g.cfg.SecretKey), nil
	})
	if err != nil {
		return nil, ErrParsingToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrGetClaims
	}

	return claims, nil
}
