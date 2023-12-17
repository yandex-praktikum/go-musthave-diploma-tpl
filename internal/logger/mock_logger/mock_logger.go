package mock_logger

import "fmt"

type MockLogger struct {
}

func New() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Infoln(args ...interface{}) {
	fmt.Println(args...)
}

func (m *MockLogger) Errorln(args ...interface{}) {
	fmt.Println(args...)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	k := keysAndValues[:]
	k = append(k, msg)
	fmt.Println(k...)
}
