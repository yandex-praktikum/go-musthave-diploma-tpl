package logger

import (
	"go.uber.org/zap"
)

// Logger is the global zap logger; it is a no-op until InitLogger is called.
var Logger *zap.Logger = zap.NewNop()

// InitLogger creates a production zap logger at the given level.
func InitLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}

	Logger = zapLogger
	return nil
}
