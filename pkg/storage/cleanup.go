package storage

import (
	"time"

	"go.uber.org/zap"
)

// RevokeSession отзывает сессию
func (s *SessionStore) RevokeSession(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists || session.IsRevoked {
		return false
	}

	session.IsRevoked = true
	s.dirty = true

	s.logger.Info("Session revoked",
		zap.String("session_id", sessionID),
		zap.Int("user_id", session.UserID),
		zap.Time("created_at", session.CreatedAt),
		zap.Time("last_used", session.LastUsedAt),
	)

	return true
}

// RevokeAllUserSessions отзывает все сессии пользователя
func (s *SessionStore) RevokeAllUserSessions(userID int) int {
	if userID <= 0 {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	revokedCount := 0
	for sessionID, session := range s.sessions {
		if session.UserID == userID && !session.IsRevoked {
			session.IsRevoked = true
			revokedCount++
			s.logger.Debug("User session revoked",
				zap.String("session_id", sessionID),
				zap.Int("user_id", userID),
			)
		}
	}

	if revokedCount > 0 {
		s.dirty = true
		s.logger.Info("All user sessions revoked",
			zap.Int("user_id", userID),
			zap.Int("revoked_count", revokedCount),
			zap.Int("total_sessions", len(s.sessions)),
		)
	}

	return revokedCount
}

// RevokeSessionsByAge отзывает старые сессии
func (s *SessionStore) RevokeSessionsByAge(maxAge time.Duration) int {
	if maxAge <= 0 {
		maxAge = s.sessionMaxAge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	revokedCount := 0
	changed := false

	for sessionID, session := range s.sessions {
		if session.LastUsedAt.Before(cutoff) && !session.IsRevoked {
			session.IsRevoked = true
			revokedCount++
			changed = true
			s.logger.Debug("Old session revoked",
				zap.String("session_id", sessionID),
				zap.Int("user_id", session.UserID),
				zap.Duration("age", time.Since(session.LastUsedAt)),
			)
		}
	}

	if changed {
		s.dirty = true
	}

	if revokedCount > 0 {
		s.logger.Info("Old sessions revoked",
			zap.Int("revoked_count", revokedCount),
			zap.Duration("max_age", maxAge),
		)
	}

	return revokedCount
}

// Cleanup удаляет старые и отозванные сессии
func (s *SessionStore) Cleanup(maxAge time.Duration) int {
	if maxAge <= 0 {
		maxAge = s.sessionMaxAge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	deletedCount := 0
	changed := false

	for sessionID, session := range s.sessions {
		if session.LastUsedAt.Before(cutoff) || session.IsRevoked {
			delete(s.sessions, sessionID)
			deletedCount++
			changed = true
			s.logger.Debug("Session cleaned up",
				zap.String("session_id", sessionID),
				zap.Int("user_id", session.UserID),
				zap.Bool("was_revoked", session.IsRevoked),
				zap.Duration("age", time.Since(session.LastUsedAt)),
			)
		}
	}

	if changed {
		s.dirty = true
	}

	if deletedCount > 0 {
		s.logger.Info("Session cleanup completed",
			zap.Int("deleted_count", deletedCount),
			zap.Int("remaining_sessions", len(s.sessions)),
			zap.Duration("max_age", maxAge),
		)
	}

	return deletedCount
}

// cleanupWorker периодически очищает старые сессии
func (s *SessionStore) cleanupWorker() {
	s.cleanupWorkerWithInterval(defaultCleanupInterval, s.sessionMaxAge)
}

// cleanupWorkerWithInterval периодически очищает старые сессии с настройками
func (s *SessionStore) cleanupWorkerWithInterval(interval, maxAge time.Duration) {
	s.logger.Debug("Cleanup worker started",
		zap.Duration("interval", interval),
		zap.Duration("max_age", maxAge),
	)
	defer s.logger.Debug("Cleanup worker stopped")

	for {
		select {
		case <-s.doneChan:
			return
		case <-s.cleanupTicker.C:
			if s.cleanupEnabled {
				deleted := s.Cleanup(maxAge)
				if deleted > 0 {
					s.logger.Debug("Periodic cleanup completed",
						zap.Int("deleted_sessions", deleted),
						zap.Duration("interval", interval),
					)
				}
			}
		}
	}
}
