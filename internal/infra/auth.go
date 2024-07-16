package infra

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os/user"
	"strings"
	"time"
)

var UserNotFound = fmt.Errorf("user not found")
var ErrInvalidPassword = fmt.Errorf("invalid password")

type UpdateUserInfoDTO struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Birthday  string `json:"birthday" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Gender    string `json:"gender" validate:"required,oneof=M F"`
}

type UserRepository interface {
	Get(ctx context.Context, id uuid.UUID) (*user.User, error)
	Create(ctx context.Context, customerID uuid.UUID, phone string, passwordHash, passwordSalt string) (uuid.UUID, error)
	Update(ctx context.Context, userID uuid.UUID, dto UpdateUserInfoDTO) error

	//GetByPhone(ctx context.Context, phone string) (*user.User, error)
	//UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type Manager struct {
	secretKey []byte
	jwtTTL    time.Duration
	userRepo  UserRepository
}

func NewManager(
	secretKey string,
	jwtTTL time.Duration,
) *Manager {
	return &Manager{
		secretKey: []byte(secretKey),
		jwtTTL:    jwtTTL,
	}
}

func (r *Manager) NewClaims(userID uuid.UUID) *jwt.RegisteredClaims {
	claims := &jwt.RegisteredClaims{}
	claims.ID = uuid.New().String()
	claims.IssuedAt = &jwt.NumericDate{Time: time.Now()}
	claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(r.jwtTTL)}

	return claims
}

func extractRawTokenFromReq(req *http.Request) (string, error) {
	authHeaders, ok := req.Header["Authorization"]
	if !ok || len(authHeaders) == 0 {
		return "", fmt.Errorf("invalid header")
	}
	rawToken := strings.TrimPrefix(authHeaders[0], "Bearer ")

	return rawToken, nil
}

func (r *Manager) ExtractTokenFromReq(req *http.Request) (*jwt.RegisteredClaims, error) {
	rawToken, err := extractRawTokenFromReq(req)
	if err != nil {
		return nil, fmt.Errorf("no authorization")
	}
	claims := &jwt.RegisteredClaims{}
	jwtToken, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		return r.secretKey, nil
	})
	if err != nil {
		return claims, err
	}

	if !jwtToken.Valid {
		return nil, fmt.Errorf("invalid jwt token")
	}

	return claims, nil
}
func (r *Manager) GenJWT(claims *jwt.RegisteredClaims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(r.secretKey)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func (r *Manager) checkPassword(u domain.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))

	return err == nil
}
