package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/korol8484/gofermart/internal/app/domain"
	"net/http"
	"time"
)

type Service struct {
	secret     string
	expire     time.Duration
	signMethod jwt.SigningMethod
	tokenName  string
}

type claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id,omitempty"`
}

func NewJwtService(secret, tokenName string, expire time.Duration) *Service {
	return &Service{
		secret:     secret,
		expire:     expire,
		signMethod: jwt.SigningMethodHS256,
		tokenName:  tokenName,
	}
}

func (s *Service) LoadUserId(r *http.Request) (domain.UserId, error) {
	cToken, err := r.Cookie(s.tokenName)
	if err != nil {
		return 0, errors.New("user session not' start")
	}

	claim, err := s.loadClaims(cToken.Value)
	if err != nil {
		return 0, errors.New("token not valid")
	}

	return domain.UserId(claim.UserID), nil
}

// CreateSession - По хорошему надо создать структуру сессии и возвращать ее,
// для структуры методы save, read через интефрейс репозитория и там уже Cookie итд
// тут упрощенно
func (s *Service) CreateSession(w http.ResponseWriter, r *http.Request, id domain.UserId) error {
	token, err := s.buildJWTString(id)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.tokenName,
		Value:    token,
		Path:     "/",
		Secure:   false,
		HttpOnly: false,
	})

	return nil
}

func (s *Service) buildJWTString(id domain.UserId) (string, error) {
	token := jwt.NewWithClaims(s.signMethod, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expire)),
		},
		UserID: int64(id),
	})

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Service) loadClaims(tokenStr string) (*claims, error) {
	cl := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, cl,
		func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != s.signMethod.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
			}

			return []byte(s.secret), nil
		})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token not valid")
	}

	return cl, nil
}
