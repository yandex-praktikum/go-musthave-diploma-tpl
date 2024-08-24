package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Age          uint8     `json:"age"`
	Username     string    `json:"username"`
	Balance      float64   `json:"balance"`
	Withdrawn    float64   `json:"withdrawn"`
	Password     string    `json:"-"`
	RefreshToken string    `json:"-"`
}
