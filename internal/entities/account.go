package entities

import "gorm.io/gorm"

type AccountType string

const (
	AccountTypeSystemWithdraw AccountType = "system_withdraw"
	AccountTypeFree           AccountType = "free"
	AccountTypeBonus          AccountType = "bonus"
)

type Account struct {
	gorm.Model
	Type   AccountType `json:"type"`
	Sum    float32     `json:"sum"`
	UserID uint        `json:"user_id"`
	User   User        `json:"user"`
}
