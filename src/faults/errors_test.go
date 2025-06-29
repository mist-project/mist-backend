package faults_test

import (
	"context"
	"fmt"
	"log/slog"
	"mist/src/faults"
	"mist/src/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name        string
		got         *faults.CustomError
		wantMessage string
		wantCode    codes.Code
	}{
		{
			name:        "TestNotFoundError",
			got:         faults.NotFoundError("error root cause", slog.LevelDebug),
			wantMessage: faults.NotFoundMessage,
			wantCode:    codes.NotFound,
		},
		{
			name:        "TestValidationError",
			got:         faults.ValidationError("error root cause", slog.LevelDebug),
			wantMessage: faults.ValidationErrorMessage,
			wantCode:    codes.InvalidArgument,
		},
		{
			name:        "TestDatabaseError",
			got:         faults.DatabaseError("error root cause", slog.LevelDebug),
			wantMessage: faults.DatabaseErrorMessage,
			wantCode:    codes.Internal,
		},
		{
			name:        "TestAuthenticationError",
			got:         faults.AuthenticationError("error root cause", slog.LevelDebug),
			wantMessage: faults.AuthenticationErrorMessage,
			wantCode:    codes.Unauthenticated,
		},
		{
			name:        "TestAuthorizationError",
			got:         faults.AuthorizationError("error root cause", slog.LevelDebug),
			wantMessage: faults.AuthorizationErrorMessage,
			wantCode:    codes.PermissionDenied,
		},
		{
			name:        "TestUnknownError",
			got:         faults.UnknownError("error root cause", slog.LevelDebug),
			wantMessage: faults.UnknownErrorMessage,
			wantCode:    codes.Unknown,
		},
		{
			name:        "TestMarshallError",
			got:         faults.MarshallError("error root cause", slog.LevelDebug),
			wantMessage: faults.MarshallErrorMessage,
			wantCode:    codes.InvalidArgument,
		},
		{
			name:        "TestMessageProducerError",
			got:         faults.MessageProducerError("error root cause", slog.LevelDebug),
			wantMessage: faults.MessageProducerErrorMessage,
			wantCode:    codes.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ASSERT
			assert.Equal(t, tt.wantMessage, tt.got.Error())
			assert.Equal(t, tt.wantCode, tt.got.Code())
			assert.NotEmpty(t, tt.got.StackTrace())
		})
	}
}

func TestRpcCustomErrorHandler(t *testing.T) {
	var (
		requestId = "req-123"
		ctx       = context.WithValue(context.Background(), helpers.RequestIdKey, requestId)
	)

	t.Run("can_handle_custom_error_response", func(t *testing.T) {
		// ARRANGE
		ce := faults.NewError("test error", "root cause", codes.InvalidArgument, slog.LevelDebug)

		// ACT
		err := faults.RpcCustomErrorHandler(ctx, ce)

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, ce.Code(), status.Code(err))
	})

	t.Run("handles_non_custom_error", func(t *testing.T) {
		// ARRANGE
		err := fmt.Errorf("test error")
		expected := status.Errorf(codes.Unknown, "test error")

		// ACT
		result := faults.RpcCustomErrorHandler(ctx, err)

		// ASSERT
		assert.NotNil(t, result)
		assert.Equal(t, codes.Unknown, status.Code(result))
		assert.Equal(t, expected.Error(), result.Error())
	})
}
