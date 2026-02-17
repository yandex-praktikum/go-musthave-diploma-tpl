package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key"

func TestNewToken(t *testing.T) {
	tests := []struct {
		name     string
		uid      string
		duration time.Duration
		secret   string
		wantErr  bool
	}{
		{
			name:     "valid token creation",
			uid:      "user-123",
			duration: time.Hour,
			secret:   testSecret,
			wantErr:  false,
		},
		{
			name:     "valid token with long duration",
			uid:      "user-456",
			duration: 24 * time.Hour,
			secret:   testSecret,
			wantErr:  false,
		},
		{
			name:     "valid token with empty uid",
			uid:      "",
			duration: time.Hour,
			secret:   testSecret,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewToken(tt.uid, tt.duration, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("NewToken() returned empty token")
			}
		})
	}
}

func TestGetToken(t *testing.T) {
	// Создаём валидный токен
	validToken, _ := NewToken("user-123", time.Hour, testSecret)

	tests := []struct {
		name    string
		token   string
		secret  []byte
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			secret:  []byte(testSecret),
			wantErr: false,
		},
		{
			name:    "invalid token - wrong secret",
			token:   validToken,
			secret:  []byte("wrong-secret"),
			wantErr: true,
		},
		{
			name:    "invalid token - malformed",
			token:   "invalid.token.here",
			secret:  []byte(testSecret),
			wantErr: true,
		},
		{
			name:    "invalid token - empty",
			token:   "",
			secret:  []byte(testSecret),
			wantErr: true,
		},
		{
			name:    "invalid token - random string",
			token:   "randomstring",
			secret:  []byte(testSecret),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetToken(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetClaimsVerified(t *testing.T) {
	// Создаём валидный токен
	validTokenStr, _ := NewToken("user-123", time.Hour, testSecret)
	validToken, _ := GetToken(validTokenStr, []byte(testSecret))

	tests := []struct {
		name    string
		setup   func() interface{}
		wantErr bool
		wantUID string
	}{
		{
			name: "valid token claims",
			setup: func() interface{} {
				return validToken
			},
			wantErr: false,
			wantUID: "user-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			if token == nil {
				t.Skip("token is nil")
			}

			claims, err := GetClaimsVerified(validToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClaimsVerified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if uid, ok := claims["sub"].(string); !ok || uid != tt.wantUID {
					t.Errorf("GetClaimsVerified() uid = %v, want %v", uid, tt.wantUID)
				}
			}
		})
	}
}

func TestTokenRoundTrip(t *testing.T) {
	uid := "test-user-id"
	duration := time.Hour

	// Создаём токен
	tokenStr, err := NewToken(uid, duration, testSecret)
	if err != nil {
		t.Fatalf("NewToken() error = %v", err)
	}

	// Парсим токен
	token, err := GetToken(tokenStr, []byte(testSecret))
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}

	// Получаем claims
	claims, err := GetClaimsVerified(token)
	if err != nil {
		t.Fatalf("GetClaimsVerified() error = %v", err)
	}

	// Проверяем uid
	gotUID, ok := claims["sub"].(string)
	if !ok {
		t.Fatal("claims[\"sub\"] is not a string")
	}
	if gotUID != uid {
		t.Errorf("uid = %v, want %v", gotUID, uid)
	}

	// Проверяем scope
	scope, ok := claims["scope"].(string)
	if !ok {
		t.Fatal("claims[\"scope\"] is not a string")
	}
	if scope != "access" {
		t.Errorf("scope = %v, want access", scope)
	}
}

func TestExpiredToken(t *testing.T) {
	// Создаём токен с отрицательной длительностью (уже истёк)
	tokenStr, err := NewToken("user-123", -time.Hour, testSecret)
	if err != nil {
		t.Fatalf("NewToken() error = %v", err)
	}

	// Пробуем распарсить истёкший токен
	_, err = GetToken(tokenStr, []byte(testSecret))
	if err == nil {
		t.Error("GetToken() should return error for expired token")
	}
}
