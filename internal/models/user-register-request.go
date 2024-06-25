package models

type UserRegisterRequest struct {
	Login    string `json:"login" validate:"required,min=4,max=32,alphanum"`
	Password string `json:"password" validate:"required,min=4,max=32,alphanum"`
}
