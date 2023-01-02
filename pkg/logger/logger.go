package logger

import "go.uber.org/zap"

// New makes new zap.Logger by debug level.
func New(debug bool) (*zap.Logger, error) {
	if debug {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
