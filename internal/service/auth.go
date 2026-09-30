package service

// AuthManager issues and verifies authentication tokens.
type AuthManager interface {
	IssueToken(username string) (string, error)
	VerifyToken(token string) (string, error)
}
