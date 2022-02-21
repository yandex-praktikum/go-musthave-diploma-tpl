package models

type Account struct {
	ID       uint   `gorm:"primary_key"`
	Number   string `gorm:"type:varchar(150) not null"`
	Current  float64
	Withdraw float64
}
