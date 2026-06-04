package log

import (
	"log/slog"
	"os"
	"strings"

	"ai_testing/internal/config"
)

func New(cfg config.Config) *slog.Logger {
	level := parseLevel(cfg.LogLevel)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.IsDev(),
	}

	h := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(h).With("service", "api", "env", cfg.Env)
}

func parseLevel(v string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
