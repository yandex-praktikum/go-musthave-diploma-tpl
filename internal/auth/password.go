// Package auth предоставляет функции для аутентификации пользователей.
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword хеширует пароль с использованием bcrypt
// Возвращает хеш в виде строки или ошибку
func HashPassword(password string) (string, error) {
	// Используем bcrypt.DefaultCost = 10
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword проверяет соответствие пароля хешу
// Возвращает true если пароль верный, false в противном случае
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return errors.Is(err, nil)
}
