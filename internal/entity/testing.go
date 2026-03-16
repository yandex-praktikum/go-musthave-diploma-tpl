package entity

import "testing"

func TestUser(t *testing.T) *User {
	return &User{
		Login:    "user",
		Password: "password",
	}
}
