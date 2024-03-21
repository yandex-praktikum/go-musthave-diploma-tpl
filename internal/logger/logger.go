package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
}

var log *Logger

func Init(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()

	if err != nil {
		return err
	}

	log = &Logger{logger: zl}

	return nil
}

func (l *Logger) Info(title string, msg ...zapcore.Field) {

	l.logger.Info(title, msg...)
}

func (l *Logger) Warn(title string, msg ...zapcore.Field) {
	l.logger.Warn(title, msg...)
}

func (l *Logger) Fatal(title string, err ...zapcore.Field) {
	l.logger.Fatal(title, err...)
}

func (l *Logger) Error(title string, err ...zapcore.Field) {
	l.logger.Error(title, err...)
}

func Log() *Logger {
	return log
}
