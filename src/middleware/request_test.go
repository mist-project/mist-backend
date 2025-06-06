package middleware_test

import (
	"bytes"
	"context"
	"fmt"
	"mist/src/helpers"
	"mist/src/logging/logger"
	"mist/src/middleware"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type dummyRequest struct {
}

var requestIdTestHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
	return ctx, nil
}

func TestRequestLoggerInterceptor(t *testing.T) {
	t.Run("it_logs_request_details", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)
		// Context with a request ID
		ctx := context.WithValue(context.Background(), helpers.RequestIdKey, "req-123")

		// ACT
		interceptor := middleware.RequestLoggerInterceptor()
		_, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}, MockHandler)

		// ASSERT
		logOutput := buf.String()
		assert.NoError(t, err)
		assert.Contains(t, logOutput, `"request_id":"req-123"`)
		assert.Contains(t, logOutput, "/test.Service/Method")
		assert.Contains(t, logOutput, `"duration":`)
		assert.Contains(t, logOutput, `"status":"OK"`)
	})

	t.Run("it_logs_details_even_when_request_errors", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)
		// Context with a request ID
		ctx := context.WithValue(context.Background(), helpers.RequestIdKey, "req-456")

		// ACT
		interceptor := middleware.RequestLoggerInterceptor()
		_, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, status.Errorf(codes.NotFound, "%s", "boom")
		})

		// ASSERT
		s, ok := status.FromError(err)

		logOutput := buf.String()
		assert.Error(t, err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, logOutput, `"request_id":"req-456"`)
		assert.Contains(t, logOutput, `"/test.Service/Method"`)
		assert.Contains(t, logOutput, `"status":"NotFound"`)
	})

	t.Run("it_logs_details_even_with_unknown_error", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf)
		// Context with a request ID
		ctx := context.WithValue(context.Background(), helpers.RequestIdKey, "req-789")

		// ACT
		interceptor := middleware.RequestLoggerInterceptor()
		_, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("boom")
		})

		// ASSERT
		s, ok := status.FromError(err)

		logOutput := buf.String()
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, logOutput, `"request_id":"req-789"`)
		assert.Contains(t, logOutput, `"/test.Service/Method"`)
		assert.Contains(t, logOutput, `"status":"Unknown"`)
	})
}

func TestRequestIdInterceptor(t *testing.T) {
	interceptor := middleware.RequestIdInterceptor()

	t.Run("it_generates_a_new_request_id_when_there_is_no_header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// ACT
		newCtx, err := interceptor(ctx, dummyRequest{}, nil, requestIdTestHandler)

		// ASSERT
		assert.NoError(t, err)
		requestId := newCtx.(context.Context).Value(helpers.RequestIdKey)
		assert.NotNil(t, requestId, "Expected a new request ID to be generated")
	})

	t.Run("it_uses_request_id_from_header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		expectedRequestId := "test-request-id"
		headers := metadata.Pairs(helpers.RequestIdKey, expectedRequestId)
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		newCtx, err := interceptor(ctx, dummyRequest{}, nil, requestIdTestHandler)

		// ASSERT
		assert.NoError(t, err)
		requestId := newCtx.(context.Context).Value(helpers.RequestIdKey)
		assert.Equal(t, expectedRequestId, requestId, "Expected the existing request ID to be used")
	})

	t.Run("it_generates_a_new_request_id_when_there_is_no_request_in_header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		ctx = metadata.NewIncomingContext(ctx, metadata.MD{})

		// ACT
		newCtx, err := interceptor(ctx, dummyRequest{}, nil, requestIdTestHandler)

		// ASSERT
		assert.NoError(t, err)
		requestId := newCtx.(context.Context).Value(helpers.RequestIdKey)
		assert.NotNil(t, requestId, "Expected a new request ID to be generated when no request ID is present in the header")
	})
}
