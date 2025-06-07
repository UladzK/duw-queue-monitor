package logger

import (
	"log/slog"
	"strings"
)

type Config struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"`
}

func (c *Config) GetLogLevel() slog.Level {
	switch strings.ToLower(c.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
