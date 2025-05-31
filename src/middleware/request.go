package middleware

import (
	"context"
	"fmt"
	"time"

	"mist/src/logging/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const RequestIdKey string = "x-request-id"

func RequestLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var (
			err        error
			ok         bool
			statusCode codes.Code
			st         *status.Status
		)
		statusCode = codes.OK

		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		if err != nil {
			st, ok = status.FromError(err)

			if !ok {
				statusCode = codes.Unknown
			} else {
				statusCode = st.Code()
			}
		}

		logger.Info(
			logger.MessageTypeRequest,
			"request_id", ctx.Value(RequestIdKey),
			"method", info.FullMethod,
			"status", statusCode.String(),
			"duration", duration.Milliseconds(),
			"user_id", GetUserId(ctx),
		)

		return resp, err
	}
}

func RequestIdInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		headers, ok := metadata.FromIncomingContext(ctx)

		if ok {
			// Check if the request ID is already present in the headers
			// If not, generate a new one
			if requestId := headers.Get(RequestIdKey); len(requestId) > 0 {
				ctx = context.WithValue(ctx, RequestIdKey, requestId[0])
			} else {
				ctx = context.WithValue(ctx, RequestIdKey, uuid.NewString())
			}
		} else {
			// If metadata is not present, create a new request ID
			ctx = context.WithValue(ctx, RequestIdKey, uuid.NewString())
		}
		return handler(ctx, req)
	}
}

func GetRequestId(ctx context.Context) string {
	requestId, ok := ctx.Value(RequestIdKey).(string)
	if !ok || requestId == "" {
		return fmt.Sprintf("UNKNOWN-%s", uuid.NewString())
	}
	return requestId
}
