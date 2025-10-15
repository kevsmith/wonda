package cli

import (
	"log/slog"
	"os"
	"strings"
)

var logger *slog.Logger

// initLogger sets up the global logger with the specified level
func initLogger(levelStr string) {
	var level slog.Level

	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN", "WARNING":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelWarn
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)
}

// GetLogger returns the global logger
func GetLogger() *slog.Logger {
	if logger == nil {
		initLogger("WARN")
	}
	return logger
}
