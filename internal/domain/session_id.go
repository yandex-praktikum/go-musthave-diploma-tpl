package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type SessionID uuid.UUID

// ParseSessionID для явного использования в коде сервиса
func ParseSessionID(s string) (SessionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SessionID(uuid.Nil), err
	}
	return SessionID(id), nil
}

// String интерфейс стрингер
func (s *SessionID) String() string {
	return uuid.UUID(*s).String()
}

// UUID
func (s *SessionID) UUID() uuid.UUID {
	return uuid.UUID(*s)
}

// UnmarshalText позволяет Huma (и другим библиотекам)
// автоматически превращать строку из куки в тип SessionID.
func (s *SessionID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = SessionID(uuid.Nil)
		return nil
	}

	id, err := uuid.Parse(string(text))
	if err != nil {
		return fmt.Errorf("invalid session id format: %w", err)
	}

	*s = SessionID(id)
	return nil
}
