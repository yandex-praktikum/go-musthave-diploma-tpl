package entities

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	LastName   string `json:"last_name" gorm:"type:varchar"`
	FirstName  string `json:"first_name" gorm:"type:varchar"`
	MiddleName string `json:"middle_name" gorm:"type:varchar"`
	Login      string `json:"login" gorm:"type:varchar;not null;unique"`
	Password   string `json:"password" gorm:"type:varchar;not null"`
	Email      string `json:"email" gorm:"type:varchar"`
}
