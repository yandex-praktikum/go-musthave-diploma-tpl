package storage

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// SessionStore - хранилище активных сессий
type SessionStore struct {
	sessions       map[string]*SessionInfo
	mu             sync.RWMutex
	updateChan     chan string
	cleanupTicker  *time.Ticker
	saveTicker     *time.Ticker
	doneChan       chan struct{}
	logger         *zap.Logger
	cleanupEnabled bool
	persistEnabled bool
	persistFile    string
	sessionMaxAge  time.Duration
	saveMutex      sync.Mutex
	lastSaveTime   time.Time
	dirty          bool
}

// GetSession получает информацию о сессии
func (s *SessionStore) GetSession(sessionID string) (*SessionInfo, bool) {
	if sessionID == "" {
		return nil, false
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists || session.IsRevoked {
		return nil, false
	}

	s.updateLastUsedAsync(sessionID)
	return session, true
}

// GetSessionWithUpdate получает информацию о сессии и синхронно обновляет время использования
func (s *SessionStore) GetSessionWithUpdate(sessionID string) (*SessionInfo, bool) {
	if sessionID == "" {
		return nil, false
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists || session.IsRevoked {
		return nil, false
	}

	s.updateLastUsedSync(sessionID)
	return session, true
}

// CreateSession создает новую сессию
func (s *SessionStore) CreateSession(sessionID string, userID int) {
	s.CreateSessionWithMetadata(sessionID, userID, "", "")
}

// CreateSessionWithMetadata создает новую сессию с метаданными
func (s *SessionStore) CreateSessionWithMetadata(sessionID string, userID int, userAgent, ipAddress string) {
	if sessionID == "" {
		s.logger.Error("Cannot create session with empty session ID")
		return
	}

	s.mu.Lock()
	now := time.Now()
	s.sessions[sessionID] = &SessionInfo{
		UserID:     userID,
		CreatedAt:  now,
		LastUsedAt: now,
		IsRevoked:  false,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}
	s.dirty = true
	s.mu.Unlock()

	s.logger.Info("Session created",
		zap.String("session_id", sessionID),
		zap.Int("user_id", userID),
		zap.String("user_agent", userAgent),
		zap.String("ip_address", ipAddress),
		zap.Int("total_sessions", len(s.sessions)),
	)
}

// GetUserSessions возвращает все сессии пользователя
func (s *SessionStore) GetUserSessions(userID int) []*SessionInfo {
	if userID <= 0 {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []*SessionInfo
	for _, session := range s.sessions {
		if session.UserID == userID && !session.IsRevoked {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetActiveSessionsCount возвращает количество активных сессий
func (s *SessionStore) GetActiveSessionsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, session := range s.sessions {
		if !session.IsRevoked {
			count++
		}
	}
	return count
}

// GetTotalSessionsCount возвращает общее количество сессий
func (s *SessionStore) GetTotalSessionsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
