package logger_test

import (
	"testing"

	"github.com/KalessinD/gophermart/internal/logger"

	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name         string
		isProd       bool
		wantDebug    bool
		expectNilErr bool
	}{
		{
			name:         "Production mode",
			isProd:       true,
			wantDebug:    false, // zap.NewProduction по умолчанию использует InfoLevel, Debug выключен
			expectNilErr: true,
		},
		{
			name:         "Development mode",
			isProd:       false,
			wantDebug:    true, // NewConsoleLogger использует DebugLevel
			expectNilErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.NewLogger(tt.isProd)

			if (err == nil) != tt.expectNilErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, !tt.expectNilErr)
			}
			if log == nil {
				t.Fatal("NewLogger() returned nil logger")
			}

			// Проверяем настройки уровня логирования через метод Core().Enabled
			isDebugEnabled := log.Core().Enabled(zap.DebugLevel)
			if isDebugEnabled != tt.wantDebug {
				t.Errorf("Debug level enabled = %v, want %v", isDebugEnabled, tt.wantDebug)
			}

			// Проверяем, что логгер можно использовать (не паникует)
			log.Info("test log message")

			_ = log.Sync()
		})
	}
}

func TestNewConsoleLogger(t *testing.T) {
	log, err := logger.NewConsoleLogger()
	if err != nil {
		t.Fatalf("NewConsoleLogger() error = %v", err)
	}
	if log == nil {
		t.Fatal("NewConsoleLogger() returned nil logger")
	}

	if !log.Core().Enabled(zap.DebugLevel) {
		t.Error("Console logger should have DebugLevel enabled")
	}

	log.Debug("debug message works")
	_ = log.Sync()
}
