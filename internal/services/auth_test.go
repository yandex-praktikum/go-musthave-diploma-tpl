package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthService(t *testing.T) {
	jwtSecret := "test-secret"
	service := NewAuthService(jwtSecret)

	require.NotNil(t, service)
	assert.Equal(t, []byte(jwtSecret), service.jwtSecret)
}

func TestAuthService_HashPassword(t *testing.T) {
	service := NewAuthService("test-secret")
	password := "testpassword"

	hash, err := service.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestAuthService_CheckPassword(t *testing.T) {
	service := NewAuthService("test-secret")
	password := "testpassword"

	hash, err := service.HashPassword(password)
	require.NoError(t, err)

	// Проверяем правильный пароль
	err = service.CheckPassword(hash, password)
	assert.NoError(t, err)

	// Проверяем неправильный пароль
	err = service.CheckPassword(hash, "wrongpassword")
	assert.Error(t, err)
}

func TestAuthService_GenerateJWT(t *testing.T) {
	service := NewAuthService("test-secret")
	userID := int64(123)
	login := "testuser"

	token, err := service.GenerateJWT(userID, login)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_ValidateJWT(t *testing.T) {
	service := NewAuthService("test-secret")
	userID := int64(123)
	login := "testuser"

	// Генерируем токен
	token, err := service.GenerateJWT(userID, login)
	require.NoError(t, err)

	// Валидируем токен
	claims, err := service.ValidateJWT(token)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "123", claims.Subject)
	assert.Contains(t, claims.Audience, login)
}

func TestAuthService_ValidateJWT_InvalidToken(t *testing.T) {
	service := NewAuthService("test-secret")

	// Пытаемся валидировать неверный токен
	claims, err := service.ValidateJWT("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_ValidateJWT_WrongSecret(t *testing.T) {
	service1 := NewAuthService("secret1")
	service2 := NewAuthService("secret2")
	userID := int64(123)
	login := "testuser"

	// Генерируем токен с одним секретом
	token, err := service1.GenerateJWT(userID, login)
	require.NoError(t, err)

	// Пытаемся валидировать с другим секретом
	claims, err := service2.ValidateJWT(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGenerateSecret(t *testing.T) {
	secret, err := GenerateSecret()

	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.Len(t, secret, 44) // base64 encoded 32 bytes
}
