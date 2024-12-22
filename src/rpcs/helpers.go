package rpcs

import (
	"mist/src/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// func logRequestBody(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 	// Log the metadata (headers) and the request body (payload)
// 	// md, ok := metadata.FromIncomingContext(ctx)
// 	// if ok {
// 	// 	log.Printf("Metadata: %v", md)
// 	// }

// 	// // Log the request body. This assumes req implements proto.Message
// 	// log.Printf("Request Body: %v", req)

// 	// Proceed with the handler
// 	return handler(ctx, req)
// }

func ErrorHandler(err error) error {
	parsedError := service.ParseServiceError(err.Error())

	switch parsedError {
	case service.ValidationError:
		return status.Errorf(codes.InvalidArgument, "%s", err.Error())
	case service.NotFoundError:
		return status.Errorf(codes.NotFound, "%s", err.Error())
	default:
		return status.Errorf(codes.Unknown, "%s", err.Error())
	}
}
