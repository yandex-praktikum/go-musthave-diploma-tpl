package models

import "github.com/ShukinDmitriy/gophermart/internal/entities"

type UserLoginResponse struct {
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	Login      string `json:"login"`
	Email      string `json:"email"`
}

func MapUserToUserLoginResponse(user *entities.User) UserLoginResponse {
	return UserLoginResponse{
		LastName:   user.LastName,
		FirstName:  user.FirstName,
		MiddleName: user.MiddleName,
		Login:      user.Login,
		Email:      user.Email,
	}
}
