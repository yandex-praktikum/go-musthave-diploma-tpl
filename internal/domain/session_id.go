package domain

import "github.com/google/uuid"

type SessionID uuid.UUID

func (s SessionID) String() string {
	return s.String()
}
