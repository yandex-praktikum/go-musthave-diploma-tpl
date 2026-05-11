// Package logger предоставляет централизованное логирование для приложения.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Level уровень логирования
type Level = slog.Level

const (
	LevelDebug Level = slog.LevelDebug
	LevelInfo  Level = slog.LevelInfo
	LevelWarn  Level = slog.LevelWarn
	LevelError Level = slog.LevelError
)

// Config конфигурация логгера
type Config struct {
	Level     Level  // Уровень логирования
	Format    string // Формат: "json" или "text"
	AddSource bool   // Добавлять источник (файл:строка)
	Output    io.Writer
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Level:     LevelInfo,
		Format:    "text",
		AddSource: true,
		Output:    os.Stdout,
	}
}

// Logger обёртка над slog.Logger
type Logger struct {
	*slog.Logger
}

// New создаёт новый логгер
func New(cfg Config) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Форматируем время
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewDevelopment создаёт логгер для разработки
func NewDevelopment() *Logger {
	cfg := DefaultConfig()
	cfg.Level = LevelDebug
	return New(cfg)
}

// WithContext возвращает логгер с контекстом
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// slog.Logger не поддерживает контекст напрямую
	// Возвращаем тот же логгер
	return l
}

// With добавляет атрибуты к логгеру
func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(attrs...),
	}
}

// Helper функции для удобного логирования

// Debug логирует сообщение уровня DEBUG
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info логирует сообщение уровня INFO
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn логирует сообщение уровня WARN
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error логирует сообщение уровня ERROR
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// ErrorWithStack логирует ошибку со стеком
func (l *Logger) ErrorWithStack(msg string, err error, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	allArgs := append(args,
		slog.String("error", err.Error()),
		slog.String("file", file),
		slog.Int("line", line),
	)
	l.Logger.Error(msg, allArgs...)
}

// Global логгер по умолчанию
var global *Logger

func init() {
	global = NewDevelopment()
}

// Default возвращает глобальный логгер
func Default() *Logger {
	return global
}

// SetDefault устанавливает глобальный логгер
func SetDefault(l *Logger) {
	global = l
}

// Convenience функции для глобального логгера

func Debug(msg string, args ...any) {
	global.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	global.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	global.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	global.Error(msg, args...)
}

func ErrorWithStack(msg string, err error, args ...any) {
	global.ErrorWithStack(msg, err, args...)
}
