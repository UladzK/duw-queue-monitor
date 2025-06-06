package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(cfg *Config) *Logger {
	opts := &slog.HandlerOptions{
		Level: cfg.GetLogLevel(),
	}

	// TODO: add distributed logging with writing both to stdout and to a file. see: https://github.com/samber/slog-multi#broadcast-slogmultifanout
	handler := slog.NewTextHandler(os.Stdout, opts)
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
