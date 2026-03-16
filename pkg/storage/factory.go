package storage

import (
	"time"

	"go.uber.org/zap"
)

// NewSessionStore создает новое хранилище сессий с логгером
func NewSessionStore(logger *zap.Logger) *SessionStore {
	if logger == nil {
		logger = zap.NewNop()
	}

	store := &SessionStore{
		sessions:       make(map[string]*SessionInfo),
		updateChan:     make(chan string, sessionUpdateBuffer),
		cleanupTicker:  time.NewTicker(defaultCleanupInterval),
		saveTicker:     time.NewTicker(defaultSaveInterval),
		doneChan:       make(chan struct{}),
		logger:         logger,
		cleanupEnabled: true,
		persistEnabled: true,
		persistFile:    defaultPersistFile,
		sessionMaxAge:  defaultSessionMaxAge,
		lastSaveTime:   time.Now(),
	}

	// Запускаем горутины для обработки обновлений, очистки и сохранения
	go store.processUpdates()
	go store.cleanupWorker()
	go store.autoSaveWorker()

	// Пытаемся загрузить сохраненные сессии
	if err := store.LoadFromFile(); err != nil {
		logger.Warn("Failed to load sessions from file", zap.Error(err))
	} else {
		logger.Info("Sessions loaded from file",
			zap.String("file", store.persistFile),
			zap.Int("session_count", len(store.sessions)),
			zap.Duration("session_max_age", store.sessionMaxAge),
		)
	}

	logger.Info("Session store initialized",
		zap.Duration("cleanup_interval", defaultCleanupInterval),
		zap.Duration("session_max_age", store.sessionMaxAge),
		zap.Duration("save_interval", defaultSaveInterval),
		zap.String("persist_file", store.persistFile),
	)

	return store
}

// NewSessionStoreWithOptions создает хранилище сессий с настройками
func NewSessionStoreWithOptions(logger *zap.Logger, cleanupInterval, maxAge time.Duration, persistFile string) *SessionStore {
	if logger == nil {
		logger = zap.NewNop()
	}

	if cleanupInterval <= 0 {
		cleanupInterval = defaultCleanupInterval
	}
	if maxAge <= 0 {
		maxAge = defaultSessionMaxAge
	}
	if persistFile == "" {
		persistFile = defaultPersistFile
	}

	store := &SessionStore{
		sessions:       make(map[string]*SessionInfo),
		updateChan:     make(chan string, sessionUpdateBuffer),
		cleanupTicker:  time.NewTicker(cleanupInterval),
		saveTicker:     time.NewTicker(defaultSaveInterval),
		doneChan:       make(chan struct{}),
		logger:         logger,
		cleanupEnabled: true,
		persistEnabled: true,
		persistFile:    persistFile,
		sessionMaxAge:  maxAge,
		lastSaveTime:   time.Now(),
	}

	go store.processUpdates()
	go store.cleanupWorkerWithInterval(cleanupInterval, maxAge)
	go store.autoSaveWorker()

	// Пытаемся загрузить сохраненные сессии
	if err := store.LoadFromFile(); err != nil {
		logger.Warn("Failed to load sessions from file", zap.Error(err))
	} else {
		logger.Info("Sessions loaded from file",
			zap.String("file", store.persistFile),
			zap.Int("session_count", len(store.sessions)),
			zap.Duration("session_max_age", store.sessionMaxAge),
		)
	}

	logger.Info("Session store initialized with custom options",
		zap.Duration("cleanup_interval", cleanupInterval),
		zap.Duration("session_max_age", maxAge),
		zap.Duration("save_interval", defaultSaveInterval),
		zap.String("persist_file", persistFile),
	)

	return store
}
