package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/alexermolov/go-kafka-pusher/internal/config"
)

// New creates a new structured logger based on configuration
func New(cfg *config.LoggingConfig) *slog.Logger {
	level := parseLevel(cfg.Level)
	
	var handler slog.Handler
	
	opts := &slog.HandlerOptions{
		Level: level,
	}

	writer := io.Writer(os.Stdout)
	
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler)
}

// parseLevel converts string log level to slog.Level
func parseLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
