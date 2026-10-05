package model

// User is the registration and login request payload.
type User struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

// maxCredentialLength is the maximum length of a username or password in bytes;
// passwords are capped at 72 bytes because bcrypt rejects longer inputs.
const maxCredentialLength = 72

// Validate reports whether username and password are non-empty and do not exceed maxCredentialLength.
func (user User) Validate() error {
	if user.Username == "" {
		return ErrMissingUsername
	}
	if user.Password == "" {
		return ErrMissingPassword
	}
	if len(user.Username) > maxCredentialLength {
		return ErrUsernameTooLong
	}
	if len(user.Password) > maxCredentialLength {
		return ErrPasswordTooLong
	}
	return nil
}
