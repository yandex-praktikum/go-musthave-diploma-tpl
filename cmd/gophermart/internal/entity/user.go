package entity

type User struct {
	ID       string
	Login    string
	Password string
	UserBalance
}

type UserBalance struct {
	Balance float32
	Spent   float32
}

func (u *User) IsValidPassword() bool {
	return u.Password != "" && len(u.Password) > 8
}

func (u *User) IsValidLogin() bool {
	return u.Password != "" && len(u.Login) > 8
}

func (u *User) IsEqual(other User) bool {
	return u.Login == other.Login && u.Password == other.Password
}
