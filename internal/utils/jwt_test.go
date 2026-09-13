package utils

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret"

func TestGenerateAndValidateToken(t *testing.T) {
	j := NewJWTUtils(testSecret)

	token, err := j.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	claims, err := j.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("ValidateToken() UserID = %d, want 42", claims.UserID)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
		t.Error("ValidateToken() token should not be expired")
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	j := NewJWTUtils(testSecret)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"garbage", "not-a-token"},
		{"signed with different secret", func() string {
			other := NewJWTUtils("other-secret")
			tok, err := other.GenerateToken(1)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}
			return tok
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := j.ValidateToken(tt.token); err == nil {
				t.Errorf("ValidateToken(%q) expected error, got nil", tt.token)
			}
		})
	}
}

func TestValidateTokenWrongAlg(t *testing.T) {
	j := NewJWTUtils(testSecret)

	// Токен с заголовком alg=none должен быть отклонён.
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjo0Mn0."
	if _, err := j.ValidateToken(noneToken); err == nil {
		t.Error("ValidateToken() with alg=none expected error, got nil")
	}
}

func TestRefreshToken(t *testing.T) {
	j := NewJWTUtils(testSecret)

	token, err := j.GenerateToken(7)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	refreshed, err := j.RefreshToken(token)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if refreshed == "" {
		t.Fatal("RefreshToken() returned empty token")
	}

	claims, err := j.ValidateToken(refreshed)
	if err != nil {
		t.Fatalf("ValidateToken(refreshed) error = %v", err)
	}
	if claims.UserID != 7 {
		t.Errorf("RefreshToken() UserID = %d, want 7", claims.UserID)
	}
}

func TestRefreshTokenInvalid(t *testing.T) {
	j := NewJWTUtils(testSecret)
	if _, err := j.RefreshToken("invalid"); err == nil {
		t.Error("RefreshToken(invalid) expected error, got nil")
	}
}

func TestExtractUserID(t *testing.T) {
	j := NewJWTUtils(testSecret)

	token, err := j.GenerateToken(99)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	userID, err := j.ExtractUserID(token)
	if err != nil {
		t.Fatalf("ExtractUserID() error = %v", err)
	}
	if userID != 99 {
		t.Errorf("ExtractUserID() = %d, want 99", userID)
	}

	if _, err := j.ExtractUserID("broken"); err == nil {
		t.Error("ExtractUserID(broken) expected error, got nil")
	}
}

func TestGenerateTokenDifferentUsers(t *testing.T) {
	j := NewJWTUtils(testSecret)

	tok1, _ := j.GenerateToken(1)
	tok2, _ := j.GenerateToken(2)
	if strings.EqualFold(tok1, tok2) {
		t.Error("tokens for different users should differ")
	}
}
