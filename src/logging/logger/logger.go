package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	MessageTypeRequest = "REQUEST"
	MessageTypeError   = "ERROR"
)

func InitializeLogger() {
	slog.SetDefault(slog.New(DefaultHandler(os.Stdout)))
}

func SetLogOutput(w io.Writer) {
	slog.SetDefault(slog.New(DefaultHandler(w)))
}

func DefaultHandler(w io.Writer) *slog.JSONHandler {
	if w == nil {
		w = os.Stdout
	}

	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: GetLogLevel(),
	})
}

func GetLogLevel() slog.Level {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "warn" // Default log level
	}

	switch strings.ToLower(logLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn // Fallback to WARN if an unknown level is set
	}
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}
