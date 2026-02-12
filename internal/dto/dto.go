package dto

import "time"

type UserCredential struct {
	Login    string `json: "login"`
	Password string `json: "password"`
}

type UserData struct {
	Uuid     string `json: "uuid"`
	Login    string `json: "login"`
	Password string `json: "password"`
}

func (c *UserCredential) ToUserData(uuid string) *UserData {
	return &UserData{
		Uuid:     uuid,
		Login:    c.Login,
		Password: c.Password,
	}
}

type OrderInfo struct {
	Number     string    `json: "number"`
	Status     string    `json: "status"`
	Accrual    int       `json: "accrual"`
	UploadedAt time.Time `json: "uploaded_at"`
}

type BalanceInfo struct {
	Current  float64 `json: "current"`
	Withdraw int     `json: "withdrawn"`
}

type WithdrawRequest struct {
	Order string `json: "order"`
	Sum   int    `json: "sum"`
}

type WithdrawInfo struct {
	Order       string    `json: "order"`
	Sum         int       `json: "sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
