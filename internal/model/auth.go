package model

type UserRegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRegisterResponse struct {
	Token  string `json:"access_token"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Token string `json:"access_token"`
}
