package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Age          uint8     `json:"age"`
	Username     string    `json:"username"`
	Balance      uint      `json:"balance"`
	Withdrawn    uint      `json:"withdrawn"`
	Password     string    `json:"-"`
	RefreshToken string    `json:"-"`
}
