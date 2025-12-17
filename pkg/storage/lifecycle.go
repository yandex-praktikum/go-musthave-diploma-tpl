package storage

import (
	"time"

	"go.uber.org/zap"
)

// Stop останавливает все фоновые процессы
func (s *SessionStore) Stop() {
	close(s.doneChan)
	s.cleanupTicker.Stop()
	s.saveTicker.Stop()

	if s.persistEnabled {
		if err := s.SaveToFile(); err != nil {
			s.logger.Error("Failed to save sessions on shutdown", zap.Error(err))
		} else {
			s.logger.Info("Sessions saved on shutdown")
		}
	}

	s.logger.Info("Session store stopped",
		zap.Int("active_sessions", s.GetActiveSessionsCount()),
		zap.Int("total_sessions", s.GetTotalSessionsCount()),
		zap.Bool("data_saved", s.persistEnabled && !s.dirty),
	)
}

// EnableCleanup включает/выключает автоматическую очистку
func (s *SessionStore) EnableCleanup(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupEnabled = enabled

	if enabled {
		s.cleanupTicker.Reset(defaultCleanupInterval)
	} else {
		s.cleanupTicker.Stop()
	}

	s.logger.Info("Cleanup enabled", zap.Bool("enabled", enabled))
}

// SetSessionMaxAge устанавливает максимальный возраст сессии
func (s *SessionStore) SetSessionMaxAge(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if maxAge <= 0 {
		maxAge = defaultSessionMaxAge
	}

	s.sessionMaxAge = maxAge
	s.dirty = true

	s.logger.Info("Session max age updated",
		zap.Duration("max_age", maxAge),
	)
}

// GetSessionMaxAge возвращает максимальный возраст сессии
func (s *SessionStore) GetSessionMaxAge() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionMaxAge
}

// updateLastUsedAsync асинхронно обновляет время последнего использования
func (s *SessionStore) updateLastUsedAsync(sessionID string) {
	select {
	case s.updateChan <- sessionID:
		// Успешно отправлено в канал
	default:
		s.logger.Warn("Update channel is full, skipping session update",
			zap.String("session_id", sessionID),
		)
	}
}

// updateLastUsedSync синхронно обновляет время последнего использования
func (s *SessionStore) updateLastUsedSync(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		session.LastUsedAt = time.Now()
		s.dirty = true
	}
}

// processUpdates обрабатывает обновления времени использования сессий
func (s *SessionStore) processUpdates() {
	s.logger.Debug("Session update processor started")
	defer s.logger.Debug("Session update processor stopped")

	for {
		select {
		case <-s.doneChan:
			return
		case sessionID := <-s.updateChan:
			if sessionID == "" {
				continue
			}

			s.mu.Lock()
			if session, exists := s.sessions[sessionID]; exists {
				session.LastUsedAt = time.Now()
				s.dirty = true

				if s.logger.Core().Enabled(zap.DebugLevel) {
					s.logger.Debug("Session last used updated",
						zap.String("session_id", sessionID),
						zap.Time("last_used_at", session.LastUsedAt),
					)
				}
			}
			s.mu.Unlock()
		}
	}
}

// autoSaveWorker периодически сохраняет сессии
func (s *SessionStore) autoSaveWorker() {
	s.logger.Debug("Auto-save worker started",
		zap.Duration("interval", defaultSaveInterval),
	)
	defer s.logger.Debug("Auto-save worker stopped")

	for {
		select {
		case <-s.doneChan:
			return
		case <-s.saveTicker.C:
			if s.dirty && s.persistEnabled {
				if err := s.SaveToFile(); err != nil {
					s.logger.Error("Failed to auto-save sessions", zap.Error(err))
				} else {
					s.logger.Debug("Auto-save completed")
				}
			}
		}
	}
}
