package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(isProd bool) (*zap.Logger, error) {
	if isProd {
		return zap.NewProduction()
	}

	return NewConsoleLogger()
}

func NewConsoleLogger() (*zap.Logger, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()

	// Цветной уровень (если нужен)
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	return zap.New(core, zap.AddCaller()), nil
}
