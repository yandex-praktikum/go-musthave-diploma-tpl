package logger

import "go.uber.org/zap"

type Logger interface {
	Infoln(args ...interface{})
	Errorln(args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
}

func New() (Logger, func() error) {
	l, lerr := zap.NewDevelopment()
	if lerr != nil {
		panic(lerr)
	}

	sugar := *l.Sugar()
	return &sugar, l.Sync
}
