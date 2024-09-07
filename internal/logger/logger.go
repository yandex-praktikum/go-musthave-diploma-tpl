package logger

import (
	"context"
	"log/slog"
	"os"
)

const (
	defaultLevel      = LevelInfo
	defaultAddSource  = false
	defaultIsJSON     = true
	defaultSetDefault = true
)

func NewLogger(opts ...LoggerOprion) *Logger {
	config := &LoggerOptions{
		Level:      defaultLevel,
		AddSource:  defaultAddSource,
		IsJSON:     defaultIsJSON,
		SetDefault: defaultSetDefault,
	}

	for _, opt := range opts {
		opt(config)
	}

	options := &HandlerOptions{
		AddSource: config.AddSource,
		Level:     config.Level,
	}

	var h Handler = NewTextHandler(os.Stdout, options)

	if config.IsJSON {
		h = NewJSONHandler(os.Stdout, options)
	}

	logger := New(h)

	if config.SetDefault {
		SetDefault(logger)
	}

	return logger
}

type LoggerOptions struct {
	Level      Level
	AddSource  bool
	IsJSON     bool
	SetDefault bool
}

type LoggerOprion func(options *LoggerOptions)

func WithLevel(level string) LoggerOprion {
	return func(optns *LoggerOptions) {
		var lvl Level
		if err := lvl.UnmarshalText([]byte(level)); err != nil {
			lvl = LevelInfo
		}

		optns.Level = lvl
	}
}

func WithAddSource(addSource bool) LoggerOprion {
	return func(optns *LoggerOptions) {
		optns.AddSource = addSource
	}
}

func WithIsJSON(isJSON bool) LoggerOprion {
	return func(optns *LoggerOptions) {
		optns.IsJSON = isJSON
	}
}

func WithSetDefault(setDefault bool) LoggerOprion {
	return func(optns *LoggerOptions) {
		optns.SetDefault = setDefault
	}
}

func WithAttr(ctx context.Context, attrs ...Attr) *Logger {
	logger := L(ctx)
	for _, attr := range attrs {
		logger = logger.With(attr)
	}

	return logger
}

func WithDefaultAttrs(logger *Logger, attrs ...Attr) *Logger {
	for _, attr := range attrs {
		logger = logger.With(attr)
	}

	return logger
}

func L(ctx context.Context) *Logger {
	return loggerFromContext(ctx)
}

func Default() *Logger {
	return slog.Default()
}
