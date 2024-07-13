package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"runtime"
)

const writeToDisk = false

// writerHook is a hook that writes logs of specified LogLevels to specified Writer
type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

// Fire will be called when some logging function is called with current hook
// It will format log Entry to string and write it to appropriate writer
func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		_, err = w.Write([]byte(line))
	}
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry

func Init() {
	l := logrus.New()

	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
		DisableColors: false,
		FullTimestamp: true,
	}

	if writeToDisk {
		err := os.MkdirAll("logs", 0755)

		if err != nil || os.IsExist(err) {
			panic("can't create log dir. no configured logging to files")
		} else {
			allFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
			if err != nil {
				panic(fmt.Sprintf("[Error]: %s", err))
			}

			l.SetOutput(io.Discard) // Send all logs to nowhere by default

			l.AddHook(&writerHook{
				Writer:    []io.Writer{allFile, os.Stdout},
				LogLevels: logrus.AllLevels,
			})
		}
	} else {
		l.SetOutput(os.Stdout)
	}

	l.SetLevel(logrus.TraceLevel)

	e = logrus.NewEntry(l)
}
