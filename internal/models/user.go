package models

import "github.com/google/uuid"

type User struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Age            uint8     `json:"age"`
	Email          string    `json:"email"`
	EmailConfirmed bool      `json:"email_confirmed"`
	Balance        uint      `json:"balance"`
	Withdrawn      uint      `json:"withdrawn"`
	Password       string    `json:"-"`
	RefreshToken   string    `json:"-"`
}
