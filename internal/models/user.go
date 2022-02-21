package models

type User struct {
	ID             uint   `gorm:"primary_key" json:"id,omitempty" `
	Login          string `gorm:"type:varchar(150) not null unique" json:"login" binding:"required" valid:"alphanum"`
	Password       string `gorm:"type:varchar(150) not null" json:"password" binding:"required"`
	LoyaltyAccount uint   `gorm:"not null" json:"account_number,omitempty"`
}
