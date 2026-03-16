package domain

import "time"

type ID struct {
	ID uint64
}

type Login struct {
	Login    string
	Password string
}

type Auth struct {
	ID     ID
	UserID ID
	Login  Login
}

type User struct {
	ID   ID
	Auth Auth
}

type SessionInfo struct {
	ID        ID
	CreatedAt time.Time
	UserID    ID
	Device    string
	// какая-то еще мета информация
}

type SessionToken struct {
	Token string
}
