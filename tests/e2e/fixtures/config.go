package fixtures

import (
	"path/filepath"
	"time"

	"github.com/KalessinD/gophermart/internal/config"
)

func NewTestConfig(connStr, tempDir string) *config.GophermartConfig {
	return &config.GophermartConfig{
		ListenAddr:               ":9081",
		ProcessingTimeout:        60 * time.Second,
		ReadTimeout:              5 * time.Second,
		ReadHeaderTimeout:        5 * time.Second,
		WriteTimeout:             10 * time.Second,
		IdleTimeout:              30 * time.Second,
		GracefullShutdownTimeout: 5 * time.Second,
		PsqlDSN:                  connStr,
		EncryptionKey:            "test-secret-key-for-e2e",
		AccrualAddress:           "", // Не тестируем внешний accrual
		QueueBufSize:             4,
		QueueWorkers:             10,
		AccrualClientTImeout:     3 * time.Second,
		WorkerPoolChanBuffer:     32,
		DumperStoragePath:        filepath.Join(tempDir, "server.dump"),
	}
}
