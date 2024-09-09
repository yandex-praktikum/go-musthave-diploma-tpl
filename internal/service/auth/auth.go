package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	db2 "github.com/kamencov/go-musthave-diploma-tpl/internal/storage/db"
	"time"

	"github.com/kamencov/go-musthave-diploma-tpl/internal/customerrors"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service"
	"github.com/kamencov/go-musthave-diploma-tpl/internal/service/auth/entity"
)

type ServiceAuth struct {
	tokenSalt    []byte
	passwordSalt []byte

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	storageUsers service.Storage
}

func NewServiceAuth(saltTKN, saltPSW []byte, storage service.Storage) *ServiceAuth {
	return &ServiceAuth{
		tokenSalt:    saltTKN,
		passwordSalt: saltPSW,

		accessTokenTTL:  24 * time.Hour,
		refreshTokenTTL: 24 * time.Hour,

		storageUsers: storage,
	}
}

func (s *ServiceAuth) RegisterUser(ctx context.Context, login, password string) error {
	err := s.storageUsers.CheckTableUserLogin(ctx, login)

	if errors.Is(err, customerrors.ErrUserAlreadyExists) {
		return customerrors.ErrUserAlreadyExists
	}

	if err != nil {
		return err
	}

	passwordHash := s.HashPassword(password)

	err = s.storageUsers.SaveTableUser(login, passwordHash)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceAuth) AuthUser(ctx context.Context, login, password string) (entity.Tokens, error) {
	passwordHash, ok := s.storageUsers.CheckTableUserPassword(ctx, login)
	if !ok {
		return entity.Tokens{}, customerrors.ErrNotFound
	}

	isPasswordTrue := s.doPasswordMatch(passwordHash, password)
	if !isPasswordTrue {
		return entity.Tokens{}, customerrors.ErrIsTruePassword
	}

	tokens, err := s.GeneratedTokens(ctx, login)
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

func (s *ServiceAuth) RefreshToken(ctx context.Context, token string) (entity.Tokens, error) {
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
	tokens, err := s.GeneratedTokens(ctx, claims.Login)
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

func (s *ServiceAuth) GeneratedTokens(ctx context.Context, login string) (entity.Tokens, error) {
	accessTokenID := uuid.NewString()

	accessToken, err := s.generateAccessToken(ctx, login)
	if err != nil {
		return entity.Tokens{}, err
	}

	// accessToken - служит доступом
	refreshToken, err := s.generateRefreshToken(ctx, accessTokenID, login)
	if err != nil {
		return entity.Tokens{}, err
	}

	// сохраняем accessToken в базу users
	err = s.storageUsers.SaveTableUserAndUpdateToken(login, accessToken)
	if err != nil {
		return entity.Tokens{}, err
	}

	return entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *ServiceAuth) generateAccessToken(ctx context.Context, login string) (string, error) {
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

func (s *ServiceAuth) generateRefreshToken(ctx context.Context, accessToken, login string) (string, error) {
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
