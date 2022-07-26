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
	OrderNumber uint
	UserID      uint
	Status      string
	Accrual     uint
}

type AccountBalance struct {
	Date        string
	UserID      uint
	OrderNumber uint
	TypeMove    string
	SumAccrual  int
	Balance     int
}
