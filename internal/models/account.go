package models

type Account struct {
	ID        uint    `gorm:"primary_key" json:"-"`
	Number    uint64  `gorm:"not null" json:"-"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
