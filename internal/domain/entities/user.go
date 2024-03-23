package entities

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Login    string
	PassHash []byte
}
