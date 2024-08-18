package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type JWTTokenService struct {
	SecretKey         string
	AccessExp         time.Duration
	RefreshExp        time.Duration
	RefreshTokenStore RefreshTokenStore
}

func NewJWTTokenService(
	SecretKey string,
	AccessExp time.Duration,
	RefreshExp time.Duration,
	RefreshTokenStore RefreshTokenStore,
) *JWTTokenService {
	return &JWTTokenService{
		SecretKey:         SecretKey,
		AccessExp:         AccessExp,
		RefreshExp:        RefreshExp,
		RefreshTokenStore: RefreshTokenStore,
	}
}

func (ts *JWTTokenService) GenerateTokens(ctx context.Context, UserID uuid.UUID) (*Tokens, error) {
	accessToken, err := ts.BuildAccessToken(ctx, UserID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := ts.BuildRefreshToken(ctx, UserID)
	if err != nil {
		return nil, err
	}
	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (ts *JWTTokenService) BuildAccessToken(ctx context.Context, UserID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ts.AccessExp)),
		},
		UserID: UserID,
	})
	tokenString, err := token.SignedString([]byte(ts.SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (ts *JWTTokenService) BuildRefreshToken(ctx context.Context, UserID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ts.RefreshExp)),
		},
		UserID: UserID,
	})
	tokenString, err := token.SignedString([]byte(ts.SecretKey))
	if err != nil {
		return "", err
	}
	ts.RefreshTokenStore.SaveRefreshToken(ctx, tokenString, &UserID)
	return tokenString, nil
}

func (ts *JWTTokenService) ValidateAccessToken(ctx context.Context, Token string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(Token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func (ts *JWTTokenService) ValidateRefreshToken(ctx context.Context, Token string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(Token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		TokenFromDB := ts.RefreshTokenStore.GetRefreshToken(ctx, &claims.UserID)
		if Token != TokenFromDB {
			return nil, fmt.Errorf("invalid token")
		}
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
