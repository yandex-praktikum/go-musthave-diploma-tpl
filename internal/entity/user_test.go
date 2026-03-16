package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		user    func() *User
		isValid bool
	}{
		{
			name: "valid",
			user: func() *User {
				return TestUser(t)
			},
			isValid: true,
		},
		{
			name: "empty login",
			user: func() *User {
				u := TestUser(t)
				u.Login = ""
				return u
			},
			isValid: false,
		},
		{
			name: "empty password",
			user: func() *User {
				u := TestUser(t)
				u.Password = ""
				return u
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.user().Validate())
			} else {
				assert.Error(t, tc.user().Validate())
			}
		})
	}
}

func TestUser_BeforeCreate(t *testing.T) {
	u := TestUser(t)
	assert.NoError(t, u.BeforeCreate())
	assert.NotEmpty(t, u.EncryptedPassword)
}
