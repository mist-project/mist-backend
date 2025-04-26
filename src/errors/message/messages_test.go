package message_test

import (
	"errors"
	"testing"

	"mist/src/errors/message"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateError(t *testing.T) {
	// ARRANGE
	expectedSubstring := "test validation error"

	// ACT
	err := message.ValidateError(expectedSubstring)

	// ASSERT
	require.Error(t, err)
	require.Contains(t, err.Error(), message.ValidationErrorString)
	require.Contains(t, err.Error(), expectedSubstring)
}

func TestNotFoundError(t *testing.T) {
	// ARRANGE
	expectedSubstring := "test not found error"

	// ACT
	err := message.NotFoundError(expectedSubstring)

	// ASSERT
	require.Error(t, err)
	require.Contains(t, err.Error(), message.NotFoundErrorString)
	require.Contains(t, err.Error(), expectedSubstring)
}

func TestDatabaseError(t *testing.T) {
	// ARRANGE
	expectedSubstring := "test database error"

	// ACT
	err := message.DatabaseError(expectedSubstring)

	// ASSERT
	require.Error(t, err)
	require.Contains(t, err.Error(), message.DatabaseErrorString)
	require.Contains(t, err.Error(), expectedSubstring)
}

func TestUnauthenticatedError(t *testing.T) {
	// ARRANGE
	expectedSubstring := "test unauthenticated error"

	// ACT
	err := message.UnauthenticatedError(expectedSubstring)

	// ASSERT
	require.Error(t, err)
	require.Contains(t, err.Error(), message.AuthenticationErrorString)
	require.Contains(t, err.Error(), expectedSubstring)
}

func TestUnauthorizedError(t *testing.T) {
	// ARRANGE
	expectedSubstring := "test unauthorized error"

	// ACT
	err := message.UnauthorizedError(expectedSubstring)

	// ASSERT
	require.Error(t, err)
	require.Contains(t, err.Error(), message.AuthorizationErrorString)
	require.Contains(t, err.Error(), expectedSubstring)
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		expected int
	}{
		{"ValidationError", message.ValidationErrorString, message.ValidationErrorCode},
		{"NotFoundError", message.NotFoundErrorString, message.NotFoundErrorCode},
		{"DatabaseError", message.DatabaseErrorString, message.DatabaseErrorCode},
		{"AuthenticationError", message.AuthenticationErrorString, message.AuthenticationErrorCode},
		{"AuthorizationError", message.AuthorizationErrorString, message.AuthorizationErrorCode},
		{"UnknownError", "some random error", message.UnknownErrorCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			input := tt.errStr

			// ACT
			code := message.ParseError(input)

			// ASSERT
			require.Equal(t, tt.expected, code)
		})
	}
}

func TestRpcErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		inputError     error
		expectedCode   codes.Code
		expectedSubstr string
	}{
		{
			name:           "ValidationError",
			inputError:     message.ValidateError("invalid input"),
			expectedCode:   codes.InvalidArgument,
			expectedSubstr: "invalid input",
		},
		{
			name:           "NotFoundError",
			inputError:     message.NotFoundError("missing resource"),
			expectedCode:   codes.NotFound,
			expectedSubstr: "missing resource",
		},
		{
			name:           "AuthenticationError",
			inputError:     message.UnauthenticatedError("invalid token"),
			expectedCode:   codes.Unauthenticated,
			expectedSubstr: "invalid token",
		},
		{
			name:           "AuthorizationError",
			inputError:     message.UnauthorizedError("access denied"),
			expectedCode:   codes.PermissionDenied,
			expectedSubstr: "access denied",
		},
		{
			name:           "UnknownError",
			inputError:     errors.New("unexpected failure"),
			expectedCode:   codes.Unknown,
			expectedSubstr: "unexpected failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			input := tt.inputError

			// ACT
			err := message.RpcErrorHandler(input)

			// ASSERT
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
			require.Contains(t, st.Message(), tt.expectedSubstr)
		})
	}
}
