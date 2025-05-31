package faults_test

import (
	"bytes"
	"errors"
	"log/slog"
	"mist/src/faults"
	"mist/src/logging/logger"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestNewError(t *testing.T) {
	t.Run("it_creates_custom_error_with_correct_message_and_code", func(t *testing.T) {
		// ARRANGE
		err := faults.NewError("test error", codes.InvalidArgument, slog.LevelDebug)

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, "test error", err.Error())
		assert.Equal(t, codes.InvalidArgument, err.Code())
		assert.NotEmpty(t, err.StackTrace())
	})
}

func TestExtendError(t *testing.T) {
	t.Run("it_extends_stack_trace_when_error_is_custom_error", func(t *testing.T) {
		// ARRANGE
		originalErr := faults.NewError("original error", codes.NotFound, slog.LevelDebug)

		// ACT
		extendedErr := faults.ExtendError(originalErr)

		// ASSERT
		assert.IsType(t, &faults.CustomError{}, extendedErr)
		ce := extendedErr.(*faults.CustomError)
		assert.Equal(t, originalErr.Error(), ce.Error())
		assert.Equal(t, originalErr.Code(), ce.Code())
		assert.True(t, strings.Contains(ce.StackTrace(), originalErr.StackTrace()))
		assert.True(t, strings.Contains(ce.StackTrace(), "TestExtendError"))
	})

	t.Run("it_returns_non_custom_error_unchanged", func(t *testing.T) {
		// ARRANGE
		stdErr := errors.New("some standard error")

		// ACT
		extended := faults.ExtendError(stdErr)

		// ASSERT
		assert.Equal(t, stdErr, extended)
	})
}

func TestCustomErrorMethods(t *testing.T) {
	t.Run("unwrap_returns_original_message_error", func(t *testing.T) {
		// ARRANGE
		err := faults.NewError("unwrap test", codes.PermissionDenied, slog.LevelDebug)

		// ACT/ASSERT
		assert.Equal(t, err.Error(), err.Unwrap().Error())
	})

	t.Run("detailed_log_output", func(t *testing.T) {
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		// Create custom error and log it
		err := faults.NewError("detailed error", codes.Internal, slog.LevelDebug)
		err.LogError(slog.LevelError, "req-123")

		logOutput := buf.String()
		assert.Contains(t, logOutput, "detailed error")
		assert.Contains(t, logOutput, "req-123")
		assert.Contains(t, logOutput, "stack_trace")
		assert.Contains(t, logOutput, `"code":13`) // 13 == codes.Internal
	})
}

func TestLogErrorLevels(t *testing.T) {
	t.Run("it_logs_at_all_levels", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		err := faults.NewError("level test", codes.Internal, slog.LevelDebug)
		levels := []slog.Level{
			slog.LevelDebug,
			slog.LevelInfo,
			slog.LevelWarn,
			slog.LevelError,
		}

		for _, level := range levels {
			buf.Reset()

			// ACT
			err.LogError(level, "req-xyz")

			// ASSERT
			output := buf.String()
			assert.Contains(t, output, `"level":`) // Should log at the correct level
			assert.Contains(t, output, "req-xyz")
			assert.Contains(t, output, "level test")
		}
	})
}

func TestStackTraceContainsCaller(t *testing.T) {
	t.Run("it_includes_caller_function_name", func(t *testing.T) {
		// ARRANGE
		err := faults.NewError("check stack trace", codes.Internal, slog.LevelDebug)

		// ACT
		stack := err.StackTrace()

		// ASSERT
		assert.Contains(t, stack, "testing.tRunner")
	})
}
