package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/golang-jwt/jwt/v4"
	passwordValidator "github.com/wagslane/go-password-validator"
)

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . AuthStorage
type AuthStorage interface {
	// ErrLoginIsBusy, ErrServerInternal
	RegisterUser(ctx context.Context, loginData *domain.LoginData) error
	// ErrServerInternal
	GetUserData(ctx context.Context, login string) (*domain.LoginData, error)
}

func NewAuth(conf *config.GophermartConfig, authStorage AuthStorage) *auth {
	return NewAuthFN(conf, authStorage, rand.Read)
}

// Вынесено отдельно для целей тестирования
func NewAuthFN(conf *config.GophermartConfig, authStorage AuthStorage, saltFn saltFn) *auth {
	return &auth{
		saltFn:      saltFn,
		tokenExp:    conf.TokenExp,
		tokenSecret: []byte(conf.TokenSecret),
		authStorage: authStorage,
	}
}

type auth struct {
	saltFn      saltFn
	tokenExp    time.Duration
	tokenSecret []byte
	authStorage AuthStorage
}

type saltFn func(b []byte) (n int, err error)

var LoginRegexp = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]*$")

// Минимальный уровень сложности пароля
// https://github.com/wagslane/go-password-validator
const minPassEntropyBits = 60

// Регистрация пользователя
// Возвращает ошибки:
//   - domain.ErrDataFormat
//   - domain.ErrServerInternal
//   - domain.ErrLoginIsBusy
func (a *auth) Register(ctx context.Context, regData *domain.RegistrationData) error {

	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't register - logger not found in context", domain.ErrServerInternal)
		return fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	if regData == nil {
		logger.Errorw("auth.Register", "err", "data is null")
		return fmt.Errorf("null register data: %w", domain.ErrServerInternal)
	}

	login := regData.Login
	if !LoginRegexp.MatchString(login) {
		logger.Errorw("auth.Register", "err", "login is bad", "login", regData.Login)
		return fmt.Errorf("%w: login is bad", domain.ErrDataFormat)
	}

	pass := regData.Password
	err = passwordValidator.Validate(pass, minPassEntropyBits)
	if err != nil {
		logger.Errorw("auth.Register", "err", "password is too simple", "login", regData.Login)
		return fmt.Errorf("%w: login is bad", domain.ErrDataFormat)
	}

	salt := make([]byte, 16)
	_, err = a.saltFn(salt)
	if err != nil {
		logger.Errorw("auth.Register", "err", fmt.Sprintf("server error %v", err.Error()), "login", regData.Login)
		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	saltB64 := base64.URLEncoding.EncodeToString(salt)

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pass))
	hex := h.Sum(nil)
	hexB64 := base64.URLEncoding.EncodeToString(hex)

	err = a.authStorage.RegisterUser(ctx, &domain.LoginData{
		Login: login,
		Hash:  hexB64,
		Salt:  saltB64,
	})

	if err != nil {
		if errors.Is(err, domain.ErrLoginIsBusy) {
			logger.Errorw("auth.Register", "err", err.Error(), "login", regData.Login)
			return fmt.Errorf("%w: can't register user", err)
		}
		logger.Errorw("auth.Register", "err", err.Error(), "login", regData.Login)
		return fmt.Errorf("%w: can't register user: %v", domain.ErrServerInternal, err.Error())
	}

	logger.Infow("Register", "status", "ok", "login", regData.Login)
	return nil
}

// Аутентификация пользователя
// Возвращает:
//   - domain.ErrServerInternal
//   - domain.ErrDataFormat
//   - domain.ErrWrongLoginPassword
func (a *auth) Login(ctx context.Context, userData *domain.AuthentificationData) (domain.TokenString, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't authentificate - logger not found in context", domain.ErrServerInternal)
		return "", fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	login := userData.Login
	data, err := a.authStorage.GetUserData(ctx, login)
	if err != nil {
		logger.Errorw("auth.Authentificate", "err", err.Error(), "login", login)
		return "", fmt.Errorf("%w: can't authentificate user: %v", domain.ErrServerInternal, err.Error())
	}

	if data == nil {
		logger.Errorw("auth.Authentificate", "err", "data is null", "login", login)
		return "", fmt.Errorf("wrong login/password: %w", domain.ErrWrongLoginPassword)
	}

	salt, err := base64.URLEncoding.DecodeString(data.Salt)
	if err != nil {
		logger.Errorw("auth.Authentificate", "err", fmt.Sprintf("can't extract salt: %v", err.Error()), "login", login)
		return "", fmt.Errorf("%w: can't extract salt %v", domain.ErrServerInternal, err.Error())
	}

	hash, err := base64.URLEncoding.DecodeString(data.Hash)
	if err != nil {
		logger.Errorw("auth.Authentificate", "err", fmt.Sprintf("can't extract hash: %v", err.Error()), "login", login)
		return "", fmt.Errorf("%w: can't extract hash %v", domain.ErrServerInternal, err.Error())
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(userData.Password))
	hex := h.Sum(nil)

	if !bytes.Equal(hash, hex) {
		logger.Errorw("auth.Authentificate", "err", "authentification failed", "login", login)
		return "", fmt.Errorf("%w: authentification failed", domain.ErrWrongLoginPassword)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, domain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenExp)),
		},
		UserID: data.UserID,
	})

	tokenString, err := token.SignedString([]byte(a.tokenSecret))
	if err != nil {
		logger.Errorw("auth.Authentificate", "err", err.Error(), "login", login)
		return "", fmt.Errorf("%w: can't sign token %v", domain.ErrServerInternal, err.Error())
	}

	return domain.TokenString(tokenString), nil
}

// Авторизация пользователя(проверка JWT)
// Возвращает ошибки:
//   - domain.ErrServerInternal
//   - domain.ErrAuthDataIncorrect
func (a *auth) Authorize(ctx context.Context, tokenString domain.TokenString) (*domain.AuthData, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't authorize - logger not found in context", domain.ErrServerInternal)
		return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	claims := &domain.Claims{}

	_, err = jwt.ParseWithClaims(string(tokenString), claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(a.tokenSecret), nil
	})

	if err != nil {
		logger.Infow("auth.Authorize", "err", err.Error())
		return nil, fmt.Errorf("%w: authorization failed", domain.ErrAuthDataIncorrect)
	}

	return &domain.AuthData{
		UserID: claims.UserID,
	}, nil
}
