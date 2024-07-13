package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

type LoggerMock struct {
	mock.Mock
}

func (m *LoggerMock) WithFields(fields logrus.Fields) Logger {
	m.Called(fields)
	return m
}

func (m *LoggerMock) WithField(key string, value interface{}) Logger {
	m.Called(key, value)
	return m
}

func (m *LoggerMock) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *LoggerMock) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...))
}

func (m *LoggerMock) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *LoggerMock) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...))
}

func (m *LoggerMock) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *LoggerMock) Fatalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...))
}
