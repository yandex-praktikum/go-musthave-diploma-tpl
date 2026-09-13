package domain

import "time"

// User представляет пользователя системы лояльности.
type User struct {
	// ID — идентификатор пользователя.
	ID int64 `json:"id"`
	// Login — логин пользователя.
	Login string `json:"login"`
	// Password — хеш пароля, не выводится в JSON.
	Password string `json:"-"`
	// CreatedAt — время регистрации пользователя.
	CreatedAt time.Time `json:"created_at"`
}
