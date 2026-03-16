package storage

import (
	"time"
)

const (
	defaultCleanupInterval = 1 * time.Hour
	defaultSessionMaxAge   = 7 * 24 * time.Hour // 7 дней
	sessionUpdateBuffer    = 100
	defaultSaveInterval    = 5 * time.Minute // Интервал автосохранения
	defaultPersistFile     = "sessions.json" // Файл для сохранения
	storeVersion           = "1.1"
)

// ValidateSession проверяет валидность сессии
func (s *SessionStore) ValidateSession(sessionID string) bool {
	session, exists := s.GetSession(sessionID)
	return exists && session != nil && !session.IsRevoked
}

// GetSessionAge возвращает возраст сессии
func (s *SessionStore) GetSessionAge(sessionID string) (time.Duration, bool) {
	session, exists := s.GetSession(sessionID)
	if !exists || session == nil {
		return 0, false
	}
	return time.Since(session.CreatedAt), true
}

// GetSessionInactivity возвращает время неактивности сессии
func (s *SessionStore) GetSessionInactivity(sessionID string) (time.Duration, bool) {
	session, exists := s.GetSession(sessionID)
	if !exists || session == nil {
		return 0, false
	}
	return time.Since(session.LastUsedAt), true
}

// GetStats возвращает статистику по сессиям
func (s *SessionStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_sessions"] = len(s.sessions)
	stats["session_max_age"] = s.sessionMaxAge.String()

	activeCount := 0
	revokedCount := 0
	userSessions := make(map[int]int)

	for _, session := range s.sessions {
		if session.IsRevoked {
			revokedCount++
		} else {
			activeCount++
		}
		userSessions[session.UserID]++
	}

	stats["active_sessions"] = activeCount
	stats["revoked_sessions"] = revokedCount
	stats["unique_users"] = len(userSessions)

	// Найти пользователя с наибольшим количеством сессий
	maxSessions := 0
	for _, count := range userSessions {
		if count > maxSessions {
			maxSessions = count
		}
	}
	stats["max_sessions_per_user"] = maxSessions
	stats["cleanup_enabled"] = s.cleanupEnabled
	stats["persist_enabled"] = s.persistEnabled
	stats["is_dirty"] = s.dirty

	return stats
}
