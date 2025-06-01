package logger_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"mist/src/logging/logger"

	"github.com/stretchr/testify/assert"
)

func TestInitializeLogger(t *testing.T) {
	t.Run("it_initializes_logger_without_error", func(t *testing.T) {
		// ACT
		logger.InitializeLogger()

		// ASSERT
		defaultLogger := slog.Default()
		assert.NotNil(t, defaultLogger)
	})
}

func TestLoggerLevels(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "debug") // <= Make sure to enable info-level logs

	t.Run("it_logs_debug_messages", func(t *testing.T) {
		// ARRANGE

		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		// ACT
		logger.Debug("debug message", "key", "value")

		// ASSERT
		logOutput := buf.String()
		fmt.Println(logOutput, "okay")
		assert.Contains(t, logOutput, "debug message")
		assert.Contains(t, logOutput, `"key":"value"`)
		assert.Contains(t, logOutput, `"level":"DEBUG"`)
	})

	t.Run("it_logs_info_messages", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		// ACT
		logger.Info("info message", "key", "value")

		// ASSERT
		logOutput := buf.String()
		assert.Contains(t, logOutput, "info message")
		assert.Contains(t, logOutput, `"key":"value"`)
		assert.Contains(t, logOutput, `"level":"INFO"`)
	})

	t.Run("it_logs_warn_messages", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		// ACT
		logger.Warn("warn message", "key", "value")

		// ASSERT
		logOutput := buf.String()
		assert.Contains(t, logOutput, "warn message")
		assert.Contains(t, logOutput, `"key":"value"`)
		assert.Contains(t, logOutput, `"level":"WARN"`)
	})

	t.Run("it_logs_error_messages", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		// ACT
		logger.Error("error message", "key", "value")

		// ASSERT
		logOutput := buf.String()
		assert.Contains(t, logOutput, "error message")
		assert.Contains(t, logOutput, `"key":"value"`)
		assert.Contains(t, logOutput, `"level":"ERROR"`)
	})
}

func TestGetLogLevel(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "debug") // <= Make sure to enable info-level logs

	t.Run("it_defaults_to_warn_if_env_missing", func(t *testing.T) {
		// ARRANGE
		_ = os.Unsetenv("LOG_LEVEL")

		// ACT
		level := logger.GetLogLevel()

		// ASSERT
		assert.Equal(t, slog.LevelWarn, level)
	})

	t.Run("it_can_get_debug_level", func(t *testing.T) {
		// ARRANGE
		_ = os.Setenv("LOG_LEVEL", "debug")

		// ACT
		level := logger.GetLogLevel()

		// ASSERT
		assert.Equal(t, slog.LevelDebug, level)
	})

	t.Run("it_can_get_info_level", func(t *testing.T) {
		// ARRANGE
		_ = os.Setenv("LOG_LEVEL", "info")

		// ACT
		level := logger.GetLogLevel()

		// ASSERT
		assert.Equal(t, slog.LevelInfo, level)
	})

	t.Run("it_can_get_warn_level", func(t *testing.T) {
		// ARRANGE
		_ = os.Setenv("LOG_LEVEL", "warn")

		// ACT
		level := logger.GetLogLevel()

		// ASSERT
		assert.Equal(t, slog.LevelWarn, level)
	})

	t.Run("it_can_get_error_level", func(t *testing.T) {
		// ARRANGE
		_ = os.Setenv("LOG_LEVEL", "error")

		// ACT
		level := logger.GetLogLevel()

		// ASSERT
		assert.Equal(t, slog.LevelError, level)
	})

	t.Run("it_fallbacks_to_warn_for_unknown_level", func(t *testing.T) {
		// ARRANGE
		_ = os.Setenv("LOG_LEVEL", "crazy")

		// ACT
		level := logger.GetLogLevel()

		// ASSERT
		assert.Equal(t, slog.LevelWarn, level)
	})
}

func TestDefaultHandler(t *testing.T) {
	t.Run("it_does_not_error_when_writer_is_nil", func(t *testing.T) {
		handler := logger.DefaultHandler(nil)
		assert.NotNil(t, handler)

		// We can also test that the handler writes to stdout by checking its type
		// But slog.Handler itself doesn't expose writer directly
		// So we’ll test that the handler actually works
		log := slog.New(handler)
		log.Info("test message")
	})

	t.Run("it_accepts_buffer_writer", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		handler := logger.DefaultHandler(&buf)
		assert.NotNil(t, handler)
	})
}
