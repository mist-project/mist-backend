package faults

import (
	"context"
	"fmt"
	"log/slog"
	"mist/src/helpers"
	"mist/src/logging/logger"
	"runtime"

	"google.golang.org/grpc/codes"
)

type ErrorWithTrace interface {
	Error() string
	StackTrace() string
	Code() codes.Code
	DetailedError() string
	Unwrap() error
}

type CustomError struct {
	message    error
	stackTrace string
	code       codes.Code
	debugLevel slog.Level
}

func NewError(err string, root string, code codes.Code, debugLevel slog.Level) *CustomError {
	// Get information about the caller where 2 is the number of skips
	// 0 is this function
	// 1 is the caller of this function that should be an error function like ErrGenericError
	// 2 is the function that called the error function
	pc, file, line, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()

	stackTrace := fmt.Sprintf("[%s:%v] %s\n\t%s", file, line, funcName, root)

	return &CustomError{
		message:    fmt.Errorf("%s", err),
		stackTrace: stackTrace,
		code:       code,
		debugLevel: debugLevel,
	}
}

func (ce *CustomError) Error() string {
	return ce.message.Error()
}

func (ce *CustomError) Unwrap() error {
	return ce.message
}

func (ce *CustomError) LogError(ctx context.Context) {

	request_id := helpers.GetRequestId(ctx)

	args := []any{
		"request_id", request_id,
		"message", ce.message.Error(),
		"code", ce.code,
		"stack_trace", ce.stackTrace,
	}

	switch ce.debugLevel {
	case slog.LevelDebug:
		logger.Debug(logger.MessageTypeError, args...)
	case slog.LevelInfo:
		logger.Info(logger.MessageTypeError, args...)
	case slog.LevelWarn:
		logger.Warn(logger.MessageTypeError, args...)
	case slog.LevelError:
		logger.Error(logger.MessageTypeError, args...)
	}
}

func LogError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	ce, ok := err.(*CustomError)

	if !ok {
		// If the error is not a CustomError, log it as an error
		logger.Error(logger.MessageTypeError, "message", err.Error())
		return
	}

	ce.LogError(ctx)
}

func (ce *CustomError) StackTrace() string {
	return ce.stackTrace
}

func (ce *CustomError) Code() codes.Code {
	return ce.code
}

func ExtendError(err error) error {
	ce, ok := err.(*CustomError)
	if !ok {
		return err
	}

	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	return &CustomError{
		message:    ce.message,
		stackTrace: fmt.Sprintf("%s\n[%s:%v] %s", ce.stackTrace, file, line, funcName),
		code:       ce.code,
	}
}
