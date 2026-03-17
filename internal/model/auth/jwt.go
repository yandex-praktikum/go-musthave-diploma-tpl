package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"gmart/internal/domain"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

// tokenTTL := 24 * time.Hour
//
// --- HMAC ---
// hmacSecret := []byte("your-very-secure-32-byte-long-secret")
//
// tg, _ := auth.NewTokenGenerator(jwt.SigningMethodHS256, hmacSecret, tokenTTL)
// tv, _ := auth.NewTokenVerifier(jwt.SigningMethodHS256, hmacSecret)
// token, _ := tg.Generate(42)
// userID, _ := tv.Parse(token)
//
// --- RSA ---
// privKey, _ := auth.LoadRSAPrivateKey("certs/private.pem")
// pubKey, _ := auth.LoadRSAPublicKey("certs/public.pem")
//
// tg, _ := auth.NewTokenGenerator(jwt.SigningMethodRS256, privKey, tokenTTL)
// token, _ := tg.Generate(1001)
//
// tv, _ := auth.NewTokenVerifier(jwt.SigningMethodRS256, pubKey)
// userID, _ := tv.Parse(token)
//
// tv2, _ := auth.NewTokenVerifier(jwt.SigningMethodRS256, privKey)
// userID2, _ := tv2.Parse(token)

type ctxKey int

const UserID ctxKey = iota

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	jwt.RegisteredClaims
	UserID domain.UserID `json:"user_id"`
}

// --- GENERATOR ---

type TokenGenerator struct {
	signingMethod jwt.SigningMethod
	signKey       any
	ttl           time.Duration
	now           func() time.Time
}

// NewTokenGenerator возвращает генератор токенов
func NewTokenGenerator(method jwt.SigningMethod, signKey any, ttl time.Duration) (*TokenGenerator, error) {
	if ttl <= 0 {
		return nil, errors.New("invalid ttl")
	}
	if signKey == nil {
		return nil, errors.New("invalid signKey")
	}
	if k, ok := signKey.([]byte); ok && len(k) < 32 {
		return nil, errors.New("hmac secret key is too short (min 32 bytes)")
	}

	return &TokenGenerator{
		signingMethod: method,
		signKey:       signKey,
		ttl:           ttl,
		now:           time.Now,
	}, nil
}

// Generate генерирует токен
func (g *TokenGenerator) Generate(userID domain.UserID) (domain.Token, error) {
	now := g.now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(g.signingMethod, claims)
	res, err := token.SignedString(g.signKey)
	return domain.Token(res), err
}

// --- VERIFIER ---

type TokenVerifier struct {
	signingMethod jwt.SigningMethod
	verifyKey     any
	now           func() time.Time
}

// NewTokenVerifier возвращает верификатор токена
func NewTokenVerifier(method jwt.SigningMethod, verifyKey any) (*TokenVerifier, error) {
	if verifyKey == nil {
		return nil, errors.New("invalid verifyKey")
	}
	if k, ok := verifyKey.(*rsa.PrivateKey); ok {
		verifyKey = &k.PublicKey
	}

	return &TokenVerifier{
		signingMethod: method,
		verifyKey:     verifyKey,
		now:           time.Now,
	}, nil
}

func (v *TokenVerifier) ParseAuthorizationHeader(AuthorizationHeader string) (domain.UserID, error) {
	if AuthorizationHeader == "" {
		return 0, ErrInvalidToken
	}

	// Разбиваем строку по пробелу
	parts := strings.Split(AuthorizationHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return 0, ErrInvalidToken
	}

	return v.Parse(domain.Token(parts[1]))
}

// Parse парсит токен
func (v *TokenVerifier) Parse(tokenString domain.Token) (domain.UserID, error) {
	token, err := jwt.ParseWithClaims(
		string(tokenString),
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != v.signingMethod.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return v.verifyKey, nil
		},
		jwt.WithTimeFunc(v.now),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		}
		return 0, errors.Join(ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, ErrInvalidToken
}

// LoadRSAPrivateKey Загружает RSA private key из PEM файла
func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa private key: %w", err)
	}

	return privateKey, nil
}

// LoadRSAPublicKey загружает RSA public key из PEM-файла
func LoadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa public key: %w", err)
	}

	return publicKey, nil
}

func NewTokenVerifierMiddleware(api huma.API, tokenVerifier *TokenVerifier) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		userID, err := tokenVerifier.ParseAuthorizationHeader(ctx.Header("Authorization"))
		if err != nil {
			slog.Warn("auth failed",
				slog.Any("err", err),
				slog.Any("remote_addr", ctx.Context().Value(http.LocalAddrContextKey)),
			)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next(huma.WithValue(ctx, UserID, userID))
	}
}
