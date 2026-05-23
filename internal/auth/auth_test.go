package auth

import (
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	// Проверяем, что хеши одного пароля разные (из-за соли)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call error = %v", err)
	}

	if hash == hash2 {
		t.Error("HashPassword() should produce different hashes for same password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Проверяем верный пароль
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword() returned false for correct password")
	}

	// Проверяем неверный пароль
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword() returned true for wrong password")
	}

	// Проверяем пустой хеш
	if VerifyPassword(password, "") {
		t.Error("VerifyPassword() returned true for empty hash")
	}
}

func TestTokenManager_GenerateToken(t *testing.T) {
	tm := NewTokenManager("test-secret-key", time.Hour)

	token, err := tm.GenerateToken(123)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}
}

func TestTokenManager_ValidateToken(t *testing.T) {
	tm := NewTokenManager("test-secret-key", time.Hour)

	token, err := tm.GenerateToken(123)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	userID, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if userID != 123 {
		t.Errorf("ValidateToken() userID = %v, want 123", userID)
	}
}

func TestTokenManager_ValidateToken_InvalidToken(t *testing.T) {
	tm := NewTokenManager("test-secret-key", time.Hour)

	// Проверяем неверный токен
	_, err := tm.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}

	// Проверяем токен с другим ключом
	tm2 := NewTokenManager("other-secret-key", time.Hour)
	token, _ := tm.GenerateToken(123)
	_, err = tm2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for token signed with different key")
	}
}

func TestTokenManager_ValidateToken_Expired(t *testing.T) {
	// Создаём токенменеджер с очень коротким временем жизни
	tm := NewTokenManager("test-secret-key", time.Millisecond*100)

	token, err := tm.GenerateToken(123)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Ждём истечения токена
	time.Sleep(time.Millisecond * 150)

	_, err = tm.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
}

func TestGetClaims(t *testing.T) {
	tm := NewTokenManager("test-secret-key", time.Hour)

	token, err := tm.GenerateToken(123)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := tm.GetClaims(token)
	if err != nil {
		t.Fatalf("GetClaims() error = %v", err)
	}

	if claims.UserID != 123 {
		t.Errorf("GetClaims() UserID = %v, want 123", claims.UserID)
	}
}
