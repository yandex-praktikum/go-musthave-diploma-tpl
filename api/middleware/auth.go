package middleware

import (
	"errors"
	"fmt"
	"github.com/A-Kuklin/gophermart/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"

	"github.com/A-Kuklin/gophermart/internal/config"
)

var signMethod = jwt.SigningMethodHS256

var (
	errSigningMethod = errors.New("unexpected signing method")
)

type Auth struct {
	cfg    *config.Config
	logger logrus.FieldLogger
}

func NewAuth(cfg *config.Config, logger logrus.FieldLogger) *Auth {
	return &Auth{
		cfg:    cfg,
		logger: logger,
	}
}

func (a Auth) TokenAccess(gCtx *gin.Context) {
	token := gCtx.Request.Header.Get(auth.AccessTokenHeaderKey)
	if token == "" {
		gCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := checkUserToken(token, a.cfg); err != nil {
		a.logger.WithError(err).Error("Access forbidden")
		gCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	gCtx.Next()
}

func checkUserToken(uToken string, cfg *config.Config) error {
	token, err := jwt.Parse(uToken, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || method != signMethod {
			return nil, fmt.Errorf("%s: %w", token.Header["alg"], errSigningMethod)
		}
		return []byte(cfg.SecretKey), nil
	})
	switch {
	case !token.Valid:
		return fmt.Errorf("invalid token: %w", err)
	case err != nil:
		return auth.ErrParsingToken
	default:
		return nil
	}
}
