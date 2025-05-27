package middleware

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

		log.Printf("REQUEST: %s (%d) TIME: %v\n", info.FullMethod, statusCode, duration)

		return resp, err
	}
}
