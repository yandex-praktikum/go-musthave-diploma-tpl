package models

type User struct {
	ID          int    `json:"id"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	AccessToken string `json:"access_token"`
}
