package models

type User struct {
	ID       int32   `gorm:"serial primary_key"`
	Login    string  `gorm:"not null"`
	Password string  `gorm:"not null"`
	Balance  float32 `gorm:"default 0 not null"`
}
