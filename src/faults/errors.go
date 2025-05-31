package faults

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NotFoundMessage             = "Not Found"
	ValidationErrorMessage      = "Validation Error"
	DatabaseErrorMessage        = "Internal Server Error"
	AuthenticationErrorMessage  = "Unauthenticated"
	AuthorizationErrorMessage   = "Unauthorized"
	MessageProducerErrorMessage = "Message Producer Error"
	MarshallErrorMessage        = "Unprocessable Entity: Marshalling Error"
	UnknownErrorMessage         = "Internal Server Error"
)

func NotFoundError(root string, debugLevel slog.Level) *CustomError {
	return NewError(NotFoundMessage, root, codes.NotFound, debugLevel)
}

func ValidationError(root string, debugLevel slog.Level) *CustomError {
	return NewError(ValidationErrorMessage, root, codes.InvalidArgument, debugLevel)
}

func DatabaseError(root string, debugLevel slog.Level) *CustomError {
	return NewError(DatabaseErrorMessage, root, codes.Internal, debugLevel)
}

func AuthenticationError(root string, debugLevel slog.Level) *CustomError {
	return NewError(AuthenticationErrorMessage, root, codes.Unauthenticated, debugLevel)
}

func AuthorizationError(root string, debugLevel slog.Level) *CustomError {
	return NewError(AuthorizationErrorMessage, root, codes.PermissionDenied, debugLevel)
}

func UnknownError(root string, debugLevel slog.Level) *CustomError {
	return NewError(UnknownErrorMessage, root, codes.Unknown, debugLevel)
}

func MarshallError(root string, debugLevel slog.Level) *CustomError {
	return NewError(MarshallErrorMessage, root, codes.InvalidArgument, debugLevel)
}

func MessageProducerError(root string, debugLevel slog.Level) *CustomError {
	return NewError(MessageProducerErrorMessage, root, codes.Unknown, debugLevel)
}

func RpcCustomErrorHandler(ctx context.Context, err error) error {
	ce, ok := err.(*CustomError)

	if !ok {
		return status.Errorf(codes.Unknown, "%s", err.Error())
	}

	ce.LogError(ctx)
	return status.Errorf(ce.Code(), "%s", err.Error())
}
