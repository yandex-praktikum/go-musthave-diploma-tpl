package models

type Withdrawl struct {
	ID           uint    `gorm:"primary_key" json:"-"`
	Order        string  `gorm:"type:varchar(150) not null" json:"-"`
	Sum          float64 `json:"current"`
	Processed_at float64 `json:"withdrawn"`
}
