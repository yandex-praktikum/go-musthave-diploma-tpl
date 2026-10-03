package model

// User is the registration and login request payload.
type User struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

// Validate reports whether username and password are non-empty.
func (user User) Validate() error {
	if user.Username == "" {
		return ErrMissingUsername
	}
	if user.Password == "" {
		return ErrMissingPassword
	}
	return nil
}
