package asynq

import (
	"go.uber.org/zap"
)

// Logger logs message to io.Writer at various log levels.
type Logger struct {
	base *zap.SugaredLogger
}

// NewLogger creates and returns a new instance of Logger.
// Log level is set to DebugLevel by default.
func NewLogger(logs *zap.SugaredLogger) *Logger {
	return &Logger{
		logs,
	}
}

// Debug is the lowest level of logging.
// Debug logs are intended for debugging and development purposes.
func (logger *Logger) Debug(args ...interface{}) {
	logger.base.Debug(args...)
}

// Info is used for general informational log messages.
func (logger *Logger) Info(args ...interface{}) {
	logger.base.Info(args...)
}

// Warn is used for undesired but relatively expected events,
// which may indicate a problem.
func (logger *Logger) Warn(args ...interface{}) {
	logger.base.Info(args...)
}

// Error is used for undesired and unexpected events that
// the program can recover from.
func (logger *Logger) Error(args ...interface{}) {
	logger.base.Error(args...)
}

// Fatal is used for undesired and unexpected events that
// the program cannot recover from.
func (logger *Logger) Fatal(args ...interface{}) {
	logger.base.Fatal(args...)
}
