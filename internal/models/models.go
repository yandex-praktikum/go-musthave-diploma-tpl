package models

type Users struct {
	ID       int32   `gorm:"serial primary_key"`
	Login    string  `gorm:"not null"`
	Password string  `gorm:"not null"`
	Balance  float32 `gorm:"default 0"`
}

type User struct {
	Login    string `gorm:"column:login" json:"login"`
	Password string `gorm:"column:password" json:"password"`
	Balance  int    `gorm:"column:balance" json:"balance"`
}
