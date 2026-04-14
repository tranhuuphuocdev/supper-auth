package logx

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func New(serviceName, logDir, logLevel string) (*slog.Logger, func() error, error) {
	if strings.TrimSpace(serviceName) == "" {
		serviceName = "service"
	}
	if strings.TrimSpace(logDir) == "" {
		logDir = filepath.Join("logs", serviceName)
	}

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create log dir: %w", err)
	}

	filePath := filepath.Join(logDir, serviceName+".log")
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	level := parseLevel(logLevel)
	handler := slog.NewJSONHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{Level: level})
	logger := slog.New(handler).With("service", serviceName)
	slog.SetDefault(logger)

	return logger, logFile.Close, nil
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
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
