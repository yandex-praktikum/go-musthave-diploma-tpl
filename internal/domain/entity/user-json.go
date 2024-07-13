package entity

type UserRegisterJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
