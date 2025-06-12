// Package logger provides a common logger interface with a set of helpers that can be used across the application.
// It uses the standard library's slog package for structured logging.
package logger

import (
	"log/slog"
	"os"
)

// Logger is a wrapper around log library to provide a common interface for logging.
type Logger struct {
	logger *slog.Logger
}

func NewLogger(cfg *Config) *Logger {
	opts := &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	}

	// TODO: add distributed logging with writing both to stdout and to a file. see: https://github.com/samber/slog-multi#broadcast-slogmultifanout
	handler := slog.NewTextHandler(os.Stdout, opts)
	// TODO: add standard attributes like "role_name", "role_instance", "version", "environment" etc.
	return &Logger{slog.New(handler)}
}

func (log *Logger) Debug(msg string, attrs ...any) {
	log.logger.Debug(msg, attrs...)
}

func (log *Logger) Info(msg string, attrs ...any) {
	log.logger.Info(msg, attrs...)
}

func (log *Logger) Warn(msg string, attrs ...any) {
	log.logger.Warn(msg, attrs...)
}

func (log *Logger) Error(msg string, err error, attrs ...any) {
	log.logger.Error(msg, append(attrs, "error", err)...)
}
