package repository

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// ==================== Константы ====================

const (
	defaultCleanupInterval = 1 * time.Hour
	defaultSessionMaxAge   = 7 * 24 * time.Hour // 7 дней
	sessionUpdateBuffer    = 100
)

// ==================== Структуры ====================

// SessionInfo - информация о сессии
type SessionInfo struct {
	UserID     int
	CreatedAt  time.Time
	LastUsedAt time.Time
	IsRevoked  bool
	UserAgent  string // Дополнительная информация о клиенте
	IPAddress  string // IP адрес клиента
}

// SessionStore - хранилище активных сессий
type SessionStore struct {
	sessions       map[string]*SessionInfo // sessionID -> SessionInfo
	mu             sync.RWMutex
	updateChan     chan string
	cleanupTicker  *time.Ticker
	doneChan       chan struct{}
	logger         *zap.Logger
	cleanupEnabled bool
}

// ==================== Фабричные функции ====================

// NewSessionStore создает новое хранилище сессий с логгером
func NewSessionStore(logger *zap.Logger) *SessionStore {
	if logger == nil {
		logger = zap.NewNop()
	}

	store := &SessionStore{
		sessions:       make(map[string]*SessionInfo),
		updateChan:     make(chan string, sessionUpdateBuffer),
		cleanupTicker:  time.NewTicker(defaultCleanupInterval),
		doneChan:       make(chan struct{}),
		logger:         logger,
		cleanupEnabled: true,
	}

	// Запускаем горутины для обработки обновлений и очистки
	go store.processUpdates()
	go store.cleanupWorker()

	logger.Info("Session store initialized",
		zap.Duration("cleanup_interval", defaultCleanupInterval),
		zap.Duration("session_max_age", defaultSessionMaxAge),
	)

	return store
}

// NewSessionStoreWithOptions создает хранилище сессий с настройками
func NewSessionStoreWithOptions(logger *zap.Logger, cleanupInterval, maxAge time.Duration) *SessionStore {
	if logger == nil {
		logger = zap.NewNop()
	}

	if cleanupInterval <= 0 {
		cleanupInterval = defaultCleanupInterval
	}
	if maxAge <= 0 {
		maxAge = defaultSessionMaxAge
	}

	store := &SessionStore{
		sessions:       make(map[string]*SessionInfo),
		updateChan:     make(chan string, sessionUpdateBuffer),
		cleanupTicker:  time.NewTicker(cleanupInterval),
		doneChan:       make(chan struct{}),
		logger:         logger,
		cleanupEnabled: true,
	}

	go store.processUpdates()
	go store.cleanupWorkerWithInterval(cleanupInterval, maxAge)

	logger.Info("Session store initialized with custom options",
		zap.Duration("cleanup_interval", cleanupInterval),
		zap.Duration("session_max_age", maxAge),
	)

	return store
}

// ==================== Управление сессиями ====================

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
	defer s.mu.Unlock()

	now := time.Now()
	s.sessions[sessionID] = &SessionInfo{
		UserID:     userID,
		CreatedAt:  now,
		LastUsedAt: now,
		IsRevoked:  false,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}

	s.logger.Info("Session created",
		zap.String("session_id", sessionID),
		zap.Int("user_id", userID),
		zap.String("user_agent", userAgent),
		zap.String("ip_address", ipAddress),
		zap.Int("total_sessions", len(s.sessions)),
	)
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

	// Асинхронно обновляем время последнего использования
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

	// Синхронно обновляем время последнего использования
	s.updateLastUsedSync(sessionID)

	return session, true
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
	}
}

// processUpdates обрабатывает обновления времени использования сессий
func (s *SessionStore) processUpdates() {
	for {
		select {
		case <-s.doneChan:
			s.logger.Debug("Session update processor stopped")
			return
		case sessionID := <-s.updateChan:
			s.updateLastUsedSync(sessionID)
		}
	}
}

// ==================== Отзыв сессий ====================

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
		maxAge = defaultSessionMaxAge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	revokedCount := 0

	for sessionID, session := range s.sessions {
		if session.LastUsedAt.Before(cutoff) && !session.IsRevoked {
			session.IsRevoked = true
			revokedCount++
			s.logger.Debug("Old session revoked",
				zap.String("session_id", sessionID),
				zap.Int("user_id", session.UserID),
				zap.Duration("age", time.Since(session.LastUsedAt)),
			)
		}
	}

	if revokedCount > 0 {
		s.logger.Info("Old sessions revoked",
			zap.Int("revoked_count", revokedCount),
			zap.Duration("max_age", maxAge),
		)
	}

	return revokedCount
}

// ==================== Очистка и управление ====================

// Cleanup удаляет старые и отозванные сессии
func (s *SessionStore) Cleanup(maxAge time.Duration) int {
	if maxAge <= 0 {
		maxAge = defaultSessionMaxAge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	deletedCount := 0

	for sessionID, session := range s.sessions {
		if session.LastUsedAt.Before(cutoff) || session.IsRevoked {
			delete(s.sessions, sessionID)
			deletedCount++
			s.logger.Debug("Session cleaned up",
				zap.String("session_id", sessionID),
				zap.Int("user_id", session.UserID),
				zap.Bool("was_revoked", session.IsRevoked),
				zap.Duration("age", time.Since(session.LastUsedAt)),
			)
		}
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
	s.cleanupWorkerWithInterval(defaultCleanupInterval, defaultSessionMaxAge)
}

// cleanupWorkerWithInterval периодически очищает старые сессии с настройками
func (s *SessionStore) cleanupWorkerWithInterval(interval, maxAge time.Duration) {
	for {
		select {
		case <-s.doneChan:
			s.logger.Debug("Cleanup worker stopped")
			return
		case <-s.cleanupTicker.C:
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

// GetStats возвращает статистику по сессиям
func (s *SessionStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_sessions"] = len(s.sessions)

	activeCount := 0
	revokedCount := 0
	userSessions := make(map[int]int)
	now := time.Now()

	for _, session := range s.sessions {
		if session.IsRevoked {
			revokedCount++
		} else {
			activeCount++
		}
		userSessions[session.UserID]++

		// Статистика по возрасту сессий
		age := now.Sub(session.LastUsedAt)
		if age > 24*time.Hour {
			// Можно добавить бакеты для разных возрастов
		}
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

	return stats
}

// Stop останавливает все фоновые процессы
func (s *SessionStore) Stop() {
	close(s.doneChan)
	s.cleanupTicker.Stop()

	s.logger.Info("Session store stopped",
		zap.Int("active_sessions", s.GetActiveSessionsCount()),
		zap.Int("total_sessions", s.GetTotalSessionsCount()),
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

// ==================== Вспомогательные функции ====================

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
