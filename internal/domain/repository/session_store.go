package repository

import (
	"sync"
	"time"
)

// SessionInfo - информация о сессии
type SessionInfo struct {
	UserID     int
	CreatedAt  time.Time
	LastUsedAt time.Time
	IsRevoked  bool
}

// SessionStore - хранилище активных сессий
type SessionStore struct {
	sessions map[string]*SessionInfo // sessionID -> SessionInfo
	mu       sync.RWMutex
}

// NewSessionStore создает новое хранилище сессий
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*SessionInfo),
	}
}

// CreateSession создает новую сессию
func (s *SessionStore) CreateSession(sessionID string, userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.sessions[sessionID] = &SessionInfo{
		UserID:     userID,
		CreatedAt:  now,
		LastUsedAt: now,
		IsRevoked:  false,
	}
}

// GetSession получает информацию о сессии
func (s *SessionStore) GetSession(sessionID string) (*SessionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists || session.IsRevoked {
		return nil, false
	}

	// Обновляем время последнего использования
	go s.UpdateLastUsed(sessionID)

	return session, true
}

// UpdateLastUsed обновляет время последнего использования сессии
func (s *SessionStore) UpdateLastUsed(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		session.LastUsedAt = time.Now()
	}
}

// RevokeSession отзывает сессию
func (s *SessionStore) RevokeSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		session.IsRevoked = true
	}
}

// RevokeAllUserSessions отзывает все сессии пользователя
func (s *SessionStore) RevokeAllUserSessions(userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.sessions {
		if session.UserID == userID {
			session.IsRevoked = true
		}
	}
}

// Cleanup удаляет старые сессии
func (s *SessionStore) Cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, session := range s.sessions {
		if session.LastUsedAt.Before(cutoff) || session.IsRevoked {
			delete(s.sessions, id)
		}
	}
}
