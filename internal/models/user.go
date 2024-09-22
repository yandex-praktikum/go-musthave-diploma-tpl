package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"login"`
	Password string `json:"password"` // Храним хэшированный пароль
}
