package handler

type RequestRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RequestLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RequestOrder struct {
	Order string `json:"order"`
}

type BalanceWithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
