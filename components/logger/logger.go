package logger

import (
	"os"

	"github.com/rs/zerolog"
)

const logTimeFormat = "2006-01-02T15:04:05.999"

func NewLogger(level string) zerolog.Logger {
	loggerLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		loggerLevel = zerolog.InfoLevel
	}
	multi := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: logTimeFormat},
	)

	return zerolog.New(multi).Level(loggerLevel).With().Timestamp().Logger()
}

func DefaultLogger() zerolog.Logger {
	return NewLogger("info")
}
