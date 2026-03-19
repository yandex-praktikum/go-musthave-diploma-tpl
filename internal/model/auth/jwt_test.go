package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"gmart/internal/domain"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager_HMAC(t *testing.T) {
	secret := []byte("super-secret-key-that-is-at-least-32-bytes-long")
	ttl := time.Hour
	userID := domain.UserID(12345)

	tg, err := NewTokenGenerator(jwt.SigningMethodHS256, secret, ttl)
	require.NoError(t, err)

	tv, err := NewTokenVerifier(jwt.SigningMethodHS256, secret)
	require.NoError(t, err)

	t.Run("success generate and parse", func(t *testing.T) {
		token, err := tg.Generate(userID)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		parsedID, err := tv.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, userID, parsedID)
	})

	t.Run("token expired", func(t *testing.T) {
		// Фиксируем время создания
		now := time.Now()
		tg.now = func() time.Time { return now }

		token, err := tg.Generate(userID)
		require.NoError(t, err)

		// Сдвигаем время верификатора вперед на время большее TTL
		tv.now = func() time.Time { return now.Add(ttl + time.Second) }

		parsedID, err := tv.Parse(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Zero(t, parsedID)
	})

	t.Run("invalid signature", func(t *testing.T) {
		token, _ := tg.Generate(userID)

		// Создаем верификатор с другим ключом
		wrongTV, _ := NewTokenVerifier(jwt.SigningMethodHS256, []byte("wrong-secret-key-long-enough-32-bytes"))

		parsedID, err := wrongTV.Parse(token)
		assert.ErrorContains(t, err, ErrInvalidToken.Error())
		assert.Zero(t, parsedID)
	})
}

func TestTokenManager_RSA(t *testing.T) {
	// Генерируем тестовую пару ключей
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	ttl := time.Minute
	userID := domain.UserID(999)

	t.Run("success RSA flow", func(t *testing.T) {
		tg, _ := NewTokenGenerator(jwt.SigningMethodRS256, privateKey, ttl)
		tv, _ := NewTokenVerifier(jwt.SigningMethodRS256, publicKey)

		token, err := tg.Generate(userID)
		require.NoError(t, err)

		parsedID, err := tv.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, userID, parsedID)
	})

	t.Run("verifier auto-extracts public key from private", func(t *testing.T) {
		// Передаем приватный ключ в верификатор
		tv, err := NewTokenVerifier(jwt.SigningMethodRS256, privateKey)
		require.NoError(t, err)

		tg, _ := NewTokenGenerator(jwt.SigningMethodRS256, privateKey, ttl)
		token, _ := tg.Generate(userID)

		parsedID, err := tv.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, userID, parsedID)
	})
}

func TestNewTokenGenerator_Validation(t *testing.T) {
	t.Run("short hmac key", func(t *testing.T) {
		_, err := NewTokenGenerator(jwt.SigningMethodHS256, []byte("short"), time.Hour)
		assert.ErrorContains(t, err, "too short")
	})

	t.Run("nil key", func(t *testing.T) {
		_, err := NewTokenGenerator(jwt.SigningMethodHS256, nil, time.Hour)
		assert.Error(t, err)
	})
}
