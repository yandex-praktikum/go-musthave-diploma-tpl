package entity

type UserDB struct {
	ID        int
	Login     string
	Password  string
	Wallet    float64
	Withdrawn float64
}
