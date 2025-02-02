package rpcs

import (
	"mist/src/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorHandler(err error) error {
	pErr := service.ParseServiceError(err.Error())

	switch pErr {
	case service.ValidationError:
		return status.Errorf(codes.InvalidArgument, "%s", err.Error())
	case service.NotFoundError:
		return status.Errorf(codes.NotFound, "%s", err.Error())
	default:
		return status.Errorf(codes.Unknown, "%s", err.Error())
	}
}
