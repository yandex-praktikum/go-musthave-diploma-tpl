package models

type Account struct {
	ID        uint   `gorm:"primary_key"`
	Number    uint64 `gorm:"not null"`
	Current   uint64
	Withdrawn uint64
}
