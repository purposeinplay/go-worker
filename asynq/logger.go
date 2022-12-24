package asynq

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(logs *zap.SugaredLogger) *Logger {
	return &Logger{
		logs,
	}
}

func (logger *Logger) Debug(args ...interface{}) {
	logger.Debug(args...)
}

func (logger *Logger) Info(args ...interface{}) {
	logger.Info(args...)
}

func (logger *Logger) Warn(args ...interface{}) {
	logger.Info(args...)
}

func (logger *Logger) Error(args ...interface{}) {
	logger.Error(args...)
}

func (logger *Logger) Fatal(args ...interface{}) {
	logger.Fatal(args...)
}
