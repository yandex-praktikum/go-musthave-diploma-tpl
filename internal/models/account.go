package models

type Account struct {
	ID        uint    `gorm:"primary_key" json:"-"`
	Number    string  `gorm:"type:varchar(150) not null" json:"-"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
