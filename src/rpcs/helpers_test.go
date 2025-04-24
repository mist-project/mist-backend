package rpcs_test

import (
	"fmt"
	"mist/src/rpcs"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorHandler(t *testing.T) {
	// Swap out the real ParseServiceError with the mock

	tests := []struct {
		name         string
		input        error
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name:         "ValidationError",
			input:        fmt.Errorf("(-1): validation failed"),
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "(-1): validation failed",
		},
		{
			name:         "NotFoundError",
			input:        fmt.Errorf("(-2): item not found"),
			expectedCode: codes.NotFound,
			expectedMsg:  "(-2): item not found",
		},
		{
			name:         "UnknownError",
			input:        fmt.Errorf("weird DB crash"),
			expectedCode: codes.Unknown,
			expectedMsg:  "weird DB crash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rpcs.ErrorHandler(tt.input)
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", err)
			}
			if st.Code() != tt.expectedCode {
				t.Errorf("expected code %v, got %v", tt.expectedCode, st.Code())
			}
			if st.Message() != tt.expectedMsg {
				t.Errorf("expected message %q, got %q", tt.expectedMsg, st.Message())
			}
		})
	}
}
