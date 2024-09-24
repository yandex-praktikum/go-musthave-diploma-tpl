package storage

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

type Order struct {
	ID        uint      `gorm:"primaryKey"`
	Number    string    `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
	Status    string    `gorm:"not null"`
	Accrual   float64   `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	User      User      `gorm:"foreignKey:UserID"`
}

func (Order) TableName() string {
	return "orders"
}

type UserBalance struct {
	ID        uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"not null"`
	Current   float64 `gorm:"default:0"`
	Withdrawn float64 `gorm:"default:0"`
	User      User    `gorm:"foreignKey:UserID"`
}

func (UserBalance) TableName() string {
	return "user_balance"
}

type Withdrawal struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null"`
	OrderNumber string    `gorm:"not null"`
	Sum         float64   `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	User        User      `gorm:"foreignKey:UserID"`
}

func (Withdrawal) TableName() string {
	return "withdrawal"
}
