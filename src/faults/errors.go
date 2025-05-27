package errors

import (
	"fmt"
	"runtime"
)

type ErrorWithTrace interface {
	Error() string
	StackTrace() string
}

type CustomError struct {
	message    string
	stackTrace string
}

const (
	NotFoundMessage = "Not Found"
)

func (e *CustomError) Error() string {
	return e.message
}

func (e *CustomError) StackTrace() string {
	return e.stackTrace
}

func New(message string) *CustomError {
	// Get information about the caller where 2 is the number of skips
	// 0 is this function
	// 1 is the caller of this function that should be an error function like ErrGenericError
	// 2 is the function that called the error function
	pc, file, line, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()

	stackTrace := fmt.Sprintf("\t[%s:%v] %s", file, line, funcName)

	return &CustomError{
		message:    message,
		stackTrace: stackTrace,
	}
}

func ExtendError(err error) error {
	ce, ok := err.(*CustomError)
	if !ok {
		return err
	}

	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	// Append new item to the current stack trace
	ce.stackTrace += fmt.Sprintf("\n\t[%s:%v] %s", file, line, funcName)

	return ce
}

func NotFoundError() *CustomError {
	return New(NotFoundMessage)
}

// const (
// 	ValidationErrorCode     int = -1
// 	NotFoundErrorCode       int = -2
// 	DatabaseErrorCode       int = -3
// 	AuthenticationErrorCode int = -4
// 	AuthorizationErrorCode  int = -5
// 	UnknownErrorCode        int = -6
// )
