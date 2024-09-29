package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	db2 "github.com/kamencov/go-musthave-diploma-tpl/internal/storage/db"
	"time"

	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth/entity"
)

//go:generate mockgen -source=./service.go -destination=service_mock.go -package=auth
type AuthService interface {
	RegisterUser(login, password string) error
	AuthUser(login, password string) (entity.Tokens, error)
	VerifyUser(token string) (string, error)
	RefreshToken(token string) (entity.Tokens, error)
	GeneratedTokens(login string) (entity.Tokens, error)
	HashPassword(password string) string
}

type StorageAuth interface {
	CheckTableUserLogin(login string) error
	CheckTableUserPassword(password string) (string, bool)
	SaveTableUserAndUpdateToken(login, accessToken string) error
	SaveTableUser(login, passwordHash string) error
	SearchLoginByToken(accessToken string) (string, error)
}

type ServiceAuth struct {
	tokenSalt    []byte
	passwordSalt []byte

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	storage StorageAuth
}

func NewService(saltTKN, saltPSW []byte, storage StorageAuth) *ServiceAuth {
	return &ServiceAuth{
		tokenSalt:    saltTKN,
		passwordSalt: saltPSW,

		accessTokenTTL:  24 * time.Hour,
		refreshTokenTTL: 24 * time.Hour,

		storage: storage,
	}
}

func (s *ServiceAuth) RegisterUser(login, password string) error {
	err := s.storage.CheckTableUserLogin(login)

	if errors.Is(err, customerrors.ErrUserAlreadyExists) {
		return customerrors.ErrUserAlreadyExists
	}

	if err != nil {
		return err
	}

	passwordHash := s.HashPassword(password)

	err = s.storage.SaveTableUser(login, passwordHash)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceAuth) AuthUser(login, password string) (entity.Tokens, error) {
	passwordHash, ok := s.storage.CheckTableUserPassword(login)
	if !ok {
		return entity.Tokens{}, customerrors.ErrNotFound
	}

	isPasswordTrue := s.doPasswordMatch(passwordHash, password)
	if !isPasswordTrue {
		return entity.Tokens{}, customerrors.ErrIsTruePassword
	}

	tokens, err := s.GeneratedTokens(login)
	if err != nil {
		return tokens, err
	}

	return tokens, nil
}

func (s *ServiceAuth) VerifyUser(token string) (string, error) {
	claims := &entity.AccessTokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, customerrors.ErrInCorrectMethod
		}

		return s.tokenSalt, nil
	})

	if err != nil {
		return "", fmt.Errorf("incorrect token: %v", err)
	}
	if !parsedToken.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.Login, nil
}

func (s *ServiceAuth) RefreshToken(token string) (entity.Tokens, error) {
	claims := &entity.RefreshTokenClaims{}
	var storageUser db2.DateBase
	parsedRefreshToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return nil, customerrors.ErrInCorrectMethod
		}
		return s.tokenSalt, nil
	})

	if err != nil || !parsedRefreshToken.Valid {
		return entity.Tokens{}, fmt.Errorf("incorrect refresh token: %v", err)
	}

	// поиск Token в хранилище
	login, err := storageUser.SearchLoginByToken(claims.AccessTokenID)
	if err != nil || login != claims.Login {
		return entity.Tokens{}, customerrors.ErrNotFound
	}

	// валидация прошла успешно, можем генерить новую пару
	tokens, err := s.GeneratedTokens(claims.Login)
	if err != nil {
		return tokens, err
	}

	// заменяем данные о старом токене, чтобы никто не мог дважды сгенерить новую пару
	err = storageUser.SaveTableUserAndUpdateToken(login, tokens.AccessToken)
	if err != nil {
		return entity.Tokens{}, err
	}

	return tokens, nil
}

func (s *ServiceAuth) GeneratedTokens(login string) (entity.Tokens, error) {
	accessTokenID := uuid.NewString()

	accessToken, err := s.generateAccessToken(login)
	if err != nil {
		return entity.Tokens{}, err
	}

	// accessToken - служит доступом
	refreshToken, err := s.generateRefreshToken(accessTokenID, login)
	if err != nil {
		return entity.Tokens{}, err
	}

	// сохраняем accessToken в базу users
	err = s.storage.SaveTableUserAndUpdateToken(login, accessToken)
	if err != nil {
		return entity.Tokens{}, err
	}

	return entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *ServiceAuth) generateAccessToken(login string) (string, error) {
	now := time.Now()
	claims := entity.AccessTokenClaims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.tokenSalt)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return signedToken, nil
}

func (s *ServiceAuth) generateRefreshToken(accessToken, login string) (string, error) {
	now := time.Now()
	claims := entity.RefreshTokenClaims{
		Login:         login,
		AccessTokenID: accessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := refreshToken.SignedString(s.tokenSalt)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return signedToken, nil
}

func (s *ServiceAuth) HashPassword(password string) string {
	var passwordBytes = []byte(password)
	var sha512Hashes = sha256.New()

	passwordBytes = append(passwordBytes, s.passwordSalt...)

	sha512Hashes.Write(passwordBytes)

	var hashedPasswordBytes = sha512Hashes.Sum(nil)
	var hashedPasswordHax = hex.EncodeToString(hashedPasswordBytes)

	return hashedPasswordHax
}

func (s *ServiceAuth) doPasswordMatch(hashPassowrd, currPassword string) bool {
	var currPasswordHash = s.HashPassword(currPassword)

	return hashPassowrd == currPasswordHash
}
