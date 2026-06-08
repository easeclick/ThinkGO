package thinkgo

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents log severity.
type LogLevel int

const (
	LevelDebug LogLevel = -4
	LevelInfo  LogLevel = 0
	LevelWarn  LogLevel = 4
	LevelError LogLevel = 8
)

// Logger is a thin wrapper around slog.
type Logger struct {
	inner *slog.Logger
}

// NewLogger creates a new logger.
func NewLogger(level string) *Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	})

	return &Logger{
		inner: slog.New(handler),
	}
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
}

// Info logs at info level.
func (l *Logger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.inner.Warn(msg, args...)
}

// Error logs at error level.
func (l *Logger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
}

// With returns a logger with additional context.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{inner: l.inner.With(args...)}
}

// WithContext returns a logger with context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{inner: l.inner.With("request_id", ctx.Value("request_id"))}
}
