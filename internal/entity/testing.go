package entity

import "testing"

func TestUser(t *testing.T) *User {
	return &User{
		Username: "user",
		Password: "password",
	}
}
