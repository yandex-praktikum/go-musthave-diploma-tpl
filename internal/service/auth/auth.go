package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/s-lyash/go-musthave-diploma-tpl/internal/models"
)

func (s Service) Registration(user models.User) (string, error) {
	id, err := s.auth.SaveUser(user)
	if err != nil && err.Error() == "exists" {
		return "", err
	}

	cookie, err := s.createNewCookie(id)
	if err != nil {
		return "", err
	}

	return cookie, nil
}

func (s Service) Login(user models.User) (string, error) {
	id, err := s.auth.GetUserID(user)
	if err != nil {
		return "", err
	}

	cookie, err := s.createNewCookie(id)
	if err != nil {
		return "", err
	}

	return cookie, nil
}

func (s Service) FindUserByID(id string) bool {
	b := s.auth.FindUserByID(id)
	return b
}

func (s Service) createNewCookie(userID string) (string, error) {
	cookie, err := buildJWTString(userID, s.secret)
	if err != nil {
		return "", err
	}

	return cookie, nil
}

func buildJWTString(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
