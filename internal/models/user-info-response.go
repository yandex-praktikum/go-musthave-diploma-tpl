package models

type UserInfoResponse struct {
	ID         uint   `json:"id"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	Login      string `json:"login"`
	Email      string `json:"email"`
}
