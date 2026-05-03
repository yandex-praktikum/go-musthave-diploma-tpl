package models_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
)

// TestNewUser проверяет конструктор пользователя
func TestNewUser(t *testing.T) {
	login := "testuser"
	password := "securepass"
	hash := "hashed_value"
	version := 1

	user := models.NewUser(login, password, hash, version)

	if user.Login != login {
		t.Errorf("expected login %s, got %s", login, user.Login)
	}
	if user.Password != password {
		t.Errorf("expected password %s, got %s", password, user.Password)
	}
	if user.Hash != hash {
		t.Errorf("expected hash %s, got %s", hash, user.Hash)
	}
	if user.Version != version {
		t.Errorf("expected version %d, got %d", version, user.Version)
	}
}

// TestUser_Validate проверяет валидацию логина и пароля
func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantErr  error
	}{
		{
			name:     "Valid credentials",
			login:    "validLogin",
			password: "validPassword123",
			wantErr:  nil,
		},
		{
			name:     "Login exactly min length",
			login:    strings.Repeat("a", models.MinLoginLength), // 4 символа
			password: "validPass",
			wantErr:  nil,
		},
		{
			name:     "Password exactly min length",
			login:    "validLogin",
			password: strings.Repeat("b", models.MinPasswordLength), // 8 символов
			wantErr:  nil,
		},
		{
			name:     "Login too short",
			login:    "abc", // 3 символа (min 4)
			password: "validPassword",
			wantErr:  models.ErrUserWrongLoginLength,
		},
		{
			name:     "Password too short",
			login:    "validLogin",
			password: "short", // 5 символов (min 8)
			wantErr:  models.ErrUserWrongPasswordLength,
		},
		{
			name:     "Empty login",
			login:    "",
			password: "validPassword",
			wantErr:  models.ErrUserWrongLoginLength,
		},
		{
			name:     "Empty password",
			login:    "validLogin",
			password: "",
			wantErr:  models.ErrUserWrongPasswordLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				Login:    tt.login,
				Password: tt.password,
			}
			err := user.Validate()

			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUser_ToJSON проверяет сериализацию в JSON
func TestUser_ToJSON(t *testing.T) {
	// Фиксируем время для предсказуемости теста
	now := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)

	user := &models.User{
		ID:        "123",
		Login:     "testuser",
		Password:  "plain_password", // В текущей структуре он сериализуется (json:"password")
		Hash:      "secret_hash",    // Не должен попадать в JSON (json:"-")
		Version:   1,
		CreatedAt: now,
	}

	data, err := user.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() returned error: %v", err)
	}

	// Проверяем, что результат является валидным JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	// Проверяем наличие ожидаемых полей
	if result["id"] != "123" {
		t.Errorf("expected id 123, got %v", result["id"])
	}
	if result["login"] != "testuser" {
		t.Errorf("expected login testuser, got %v", result["login"])
	}

	// Проверяем, что Hash НЕ попал в JSON
	if _, ok := result["hash"]; ok {
		t.Error("Hash field should not be present in JSON, but it is")
	}

	// Проверяем, что Password попал (в коде тег json:"password", а не "-" как требует gosec)
	if _, ok := result["password"]; !ok {
		t.Error("Password field should be present in JSON based on struct tags")
	}
}

// TestFromJSON проверяет десериализацию из JSON
func TestFromJSON(t *testing.T) {
	jsonStr := `{"id":"login@domain.ru","login":"jsonuser","password":"jsonpass","version":2}`

	user, err := models.FromJSON[models.User]([]byte(jsonStr))
	if err != nil {
		t.Fatalf("FromJSON() returned error: %v", err)
	}

	if user.ID != "login@domain.ru" {
		t.Errorf("expected ID login@domain.ru, got %s", user.ID)
	}
	if user.Login != "jsonuser" {
		t.Errorf("expected Login jsonuser, got %s", user.Login)
	}
	if user.Password != "jsonpass" {
		t.Errorf("expected Password jsonpass, got %s", user.Password)
	}
	if user.Version != 2 {
		t.Errorf("expected Version 2, got %d", user.Version)
	}
}
