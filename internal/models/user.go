package models

import "time"

// User представляет пользователя системы
type User struct {
    ID        int       `json:"id"`
    Login     string    `json:"login"`
    Password  string    `json:"-"` // Поле скрыто при сериализации JSON
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// AuthRequest запрос на аутентификацию/регистрацию
type AuthRequest struct {
    Login    string `json:"login"`
    Password string `json:"password"`
}
