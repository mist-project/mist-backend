package faults_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"mist/src/faults"
	"mist/src/helpers"
	"mist/src/logging/logger"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestNewError(t *testing.T) {
	t.Run("it_creates_custom_error_with_correct_message_and_code", func(t *testing.T) {
		// ARRANGE
		err := faults.NewError("test error", "root cause", codes.InvalidArgument, slog.LevelDebug)

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
		originalErr := faults.NewError("original error", "root cause", codes.NotFound, slog.LevelDebug)

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
	var (
		requestId = "req-123"
		ctx       = context.WithValue(context.Background(), helpers.RequestIdKey, requestId)
	)
	t.Run("unwrap_returns_original_message_error", func(t *testing.T) {
		// ARRANGE
		err := faults.NewError("unwrap test", "root cause", codes.PermissionDenied, slog.LevelDebug)

		// ACT/ASSERT
		assert.Equal(t, err.Error(), err.Unwrap().Error())
	})

	t.Run("detailed_log_output", func(t *testing.T) {
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		// Create custom error and log it
		err := faults.NewError("detailed error", "root cause", codes.Internal, slog.LevelDebug)
		err.LogError(ctx)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "detailed error")
		assert.Contains(t, logOutput, requestId)
		assert.Contains(t, logOutput, "stack_trace")
		assert.Contains(t, logOutput, `"code":13`) // 13 == codes.Internal
	})
}

func TestLogErrorLevels(t *testing.T) {
	var (
		requestId = "req-123"
		ctx       = context.WithValue(context.Background(), helpers.RequestIdKey, requestId)
	)

	t.Run("it_logs_at_all_levels", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)

		levels := []slog.Level{
			slog.LevelDebug,
			slog.LevelInfo,
			slog.LevelWarn,
			slog.LevelError,
		}

		for _, level := range levels {
			err := faults.NewError("level test", "root cause", codes.Internal, level)
			buf.Reset()

			// ACT
			err.LogError(ctx)

			// ASSERT
			output := buf.String()
			assert.Contains(t, output, `"level":`) // Should log at the correct level
			assert.Contains(t, output, "root cause")
			assert.Contains(t, output, "level test")
		}
	})
}

func TestStackTraceContainsCaller(t *testing.T) {

	t.Run("it_includes_caller_function_name", func(t *testing.T) {
		// ARRANGE
		err := faults.NewError("check stack trace", "root cause", codes.Internal, slog.LevelDebug)

		// ACT
		stack := err.StackTrace()

		// ASSERT
		assert.Contains(t, stack, "testing.tRunner")
	})
}
