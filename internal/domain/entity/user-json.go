package entity

type UserRegisterJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserBalanceJSON struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
