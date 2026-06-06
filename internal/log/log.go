package log

import (
	"io"
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

	var w io.Writer = os.Stdout
	if cfg.LogSpaced {
		w = newSpacedWriter(w)
	}
	h := slog.NewJSONHandler(w, opts)
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
