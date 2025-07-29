package services

import (
	"testing"
)

func BenchmarkHashPassword(b *testing.B) {
	authService := NewAuthService("test-secret")
	password := "test-password-123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authService.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	authService := NewAuthService("test-secret")
	password := "test-password-123"
	hashedPassword, err := authService.HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := authService.CheckPassword(hashedPassword, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateJWT(b *testing.B) {
	authService := NewAuthService("test-secret")
	userID := int64(123)
	login := "testuser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authService.GenerateJWT(userID, login)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateJWT(b *testing.B) {
	authService := NewAuthService("test-secret")
	userID := int64(123)
	login := "testuser"

	token, err := authService.GenerateJWT(userID, login)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authService.ValidateJWT(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateSecret()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк для полного цикла аутентификации
func BenchmarkFullAuthCycle(b *testing.B) {
	authService := NewAuthService("test-secret")
	password := "test-password-123"
	userID := int64(123)
	login := "testuser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Хешируем пароль
		hashedPassword, err := authService.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}

		// Проверяем пароль
		err = authService.CheckPassword(hashedPassword, password)
		if err != nil {
			b.Fatal(err)
		}

		// Генерируем JWT
		token, err := authService.GenerateJWT(userID, login)
		if err != nil {
			b.Fatal(err)
		}

		// Валидируем JWT
		_, err = authService.ValidateJWT(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}
