package faults

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NotFoundMessage            = "Not Found"
	ValidationErrorMessage     = "Validation Error"
	DatabaseErrorMessage       = "Internal Server Error"
	AuthenticationErrorMessage = "Unauthenticated"
	AuthorizationErrorMessage  = "Unauthorized"
	UnknownErrorMessage        = "Internal Server Error"
)

func NotFoundError(debugLevel slog.Level) *CustomError {
	return NewError(NotFoundMessage, codes.NotFound, debugLevel)
}

func ValidationError(debugLevel slog.Level) *CustomError {
	return NewError(ValidationErrorMessage, codes.InvalidArgument, debugLevel)
}

func DatabaseError(debugLevel slog.Level) *CustomError {
	return NewError(DatabaseErrorMessage, codes.Internal, debugLevel)
}

func AuthenticationError(debugLevel slog.Level) *CustomError {
	return NewError(AuthenticationErrorMessage, codes.Unauthenticated, debugLevel)
}

func AuthorizationError(debugLevel slog.Level) *CustomError {
	return NewError(AuthorizationErrorMessage, codes.PermissionDenied, debugLevel)
}

func UnknownError(debugLevel slog.Level) *CustomError {
	return NewError(UnknownErrorMessage, codes.Unknown, debugLevel)
}

func RpcCustomErrorHandler(requestId string, err error) error {
	ce, ok := err.(*CustomError)
	if !ok {
		return status.Errorf(codes.Unknown, "%s", err.Error())
	}

	ce.LogError(ce.debugLevel, fmt.Sprintf(requestId))
	return status.Errorf(ce.Code(), "%s", err.Error())
}
