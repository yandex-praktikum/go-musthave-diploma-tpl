package storage

import (
	"time"
)

// SessionStorage интерфейс для работы с хранилищем сессий
type SessionStorage interface {
	CreateSession(sessionID string, userID int)
	CreateSessionWithMetadata(sessionID string, userID int, userAgent, ipAddress string)
	GetSession(sessionID string) (*SessionInfo, bool)
	GetSessionWithUpdate(sessionID string) (*SessionInfo, bool)
	RevokeSession(sessionID string) bool
	RevokeAllUserSessions(userID int) int
	RevokeSessionsByAge(maxAge time.Duration) int

	Cleanup(maxAge time.Duration) int
	Stop()

	GetUserSessions(userID int) []*SessionInfo
	GetActiveSessionsCount() int
	GetTotalSessionsCount() int
	GetStats() map[string]interface{}
	ValidateSession(sessionID string) bool
	GetSessionAge(sessionID string) (time.Duration, bool)
	GetSessionInactivity(sessionID string) (time.Duration, bool)

	EnablePersistence(enabled bool)
	SetPersistFile(filename string) error
	ForceSave() error
	Backup(backupDir string) (string, error)
	GetPersistInfo() map[string]interface{}
	LoadFromFile() error

	EnableCleanup(enabled bool)
	SetSessionMaxAge(maxAge time.Duration)
	GetSessionMaxAge() time.Duration
}

// SessionInfo - информация о сессии
type SessionInfo struct {
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	IsRevoked  bool      `json:"is_revoked"`
	UserAgent  string    `json:"user_agent,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
}

// PersistedData - структура для сохранения в файл
type PersistedData struct {
	Version       string                  `json:"version"`
	SavedAt       time.Time               `json:"saved_at"`
	SessionMaxAge time.Duration           `json:"session_max_age"`
	Sessions      map[string]*SessionInfo `json:"sessions"`
	TotalCount    int                     `json:"total_count"`
}
