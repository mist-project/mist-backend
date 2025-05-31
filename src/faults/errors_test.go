package faults_test

import (
	"fmt"
	"log/slog"
	"mist/src/faults"
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
			got:         faults.NotFoundError(slog.LevelDebug),
			wantMessage: faults.NotFoundMessage,
			wantCode:    codes.NotFound,
		},
		{
			name:        "TestValidationError",
			got:         faults.ValidationError(slog.LevelDebug),
			wantMessage: faults.ValidationErrorMessage,
			wantCode:    codes.InvalidArgument,
		},
		{
			name:        "TestDatabaseError",
			got:         faults.DatabaseError(slog.LevelDebug),
			wantMessage: faults.DatabaseErrorMessage,
			wantCode:    codes.Internal,
		},
		{
			name:        "TestAuthenticationError",
			got:         faults.AuthenticationError(slog.LevelDebug),
			wantMessage: faults.AuthenticationErrorMessage,
			wantCode:    codes.Unauthenticated,
		},
		{
			name:        "TestAuthorizationError",
			got:         faults.AuthorizationError(slog.LevelDebug),
			wantMessage: faults.AuthorizationErrorMessage,
			wantCode:    codes.PermissionDenied,
		},
		{
			name:        "TestUnknownError",
			got:         faults.UnknownError(slog.LevelDebug),
			wantMessage: faults.UnknownErrorMessage,
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
	t.Run("can_handle_custom_error_response", func(t *testing.T) {
		// ARRANGE
		ce := faults.NewError("test error", codes.InvalidArgument, slog.LevelDebug)
		requestId := "req-123"

		// ACT
		err := faults.RpcCustomErrorHandler(requestId, ce)

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, ce.Code(), status.Code(err))
	})

	t.Run("handles_non_custom_error", func(t *testing.T) {
		// ARRANGE
		err := fmt.Errorf("test error")
		requestId := "req-456"
		expected := status.Errorf(codes.Unknown, "test error")

		// ACT
		result := faults.RpcCustomErrorHandler(requestId, err)

		// ASSERT
		assert.NotNil(t, result)
		assert.Equal(t, codes.Unknown, status.Code(result))
		assert.Equal(t, expected.Error(), result.Error())
	})
}
