package models

type User struct {
	ID       uint
	Username string
	Password string
}

type UserAPI struct {
	Username string `json:"login"`
	Password string `json:"password"`
}
