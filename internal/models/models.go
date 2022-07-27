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

type Order struct {
	ID          uint
	Date        string
	OrderNumber string
	UserID      uint
	Status      string
	Accrual     float32
}

type AccountBalance struct {
	Date        string
	UserID      uint
	OrderNumber string
	TypeMove    string
	SumAccrual  float32
	Balance     float32
}

type AccountBalanceAPI struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type OrderAPI struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type Withdraw struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type OrderES struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}
