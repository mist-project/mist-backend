package message

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	InvalidUUID  string = "invalid uuid"
	NotFound     string = "resource not found"
	Unauthorized string = "Unauthorized"

	// DB ERRORS
	DbNotFound = "no rows in result set"
)

const (
	ValidationErrorCode     int = -1
	NotFoundErrorCode       int = -2
	DatabaseErrorCode       int = -3
	AuthenticationErrorCode int = -4
	AuthorizationErrorCode  int = -5
	UnknownErrorCode        int = -6
)

var (
	ValidationErrorString     string = fmt.Sprintf("(%d)", ValidationErrorCode)
	NotFoundErrorString       string = fmt.Sprintf("(%d)", NotFoundErrorCode)
	DatabaseErrorString       string = fmt.Sprintf("(%d)", DatabaseErrorCode)
	AuthenticationErrorString string = fmt.Sprintf("(%d)", AuthenticationErrorCode)
	AuthorizationErrorString  string = fmt.Sprintf("(%d)", AuthorizationErrorCode)
	UnknownErrorString        string = fmt.Sprintf("(%d)", UnknownErrorCode)
)

const (
	NotFoundMessage = "Not Found"
)

func ValidateError(s string) error {
	return fmt.Errorf("%s %s", ValidationErrorString, s)
}

func NotFoundError(s string) error {
	return fmt.Errorf("%s %s", NotFoundErrorString, s)
}
func DatabaseError(s string) error {
	return fmt.Errorf("%s %s", DatabaseErrorString, s)
}

func UnauthenticatedError(s string) error {
	return fmt.Errorf("%s %s", AuthenticationErrorString, s)
}

func UnauthorizedError(s string) error {
	return fmt.Errorf("%s %s", AuthorizationErrorString, s)
}

func UnknownError(s string) error {
	return fmt.Errorf("%s %s", UnknownErrorString, s)
}

func ParseError(s string) int {

	if strings.Contains(s, ValidationErrorString) {
		return ValidationErrorCode
	}

	if strings.Contains(s, NotFoundErrorString) {
		return NotFoundErrorCode
	}

	if strings.Contains(s, DatabaseErrorString) {
		return DatabaseErrorCode
	}

	if strings.Contains(s, AuthenticationErrorString) {
		return AuthenticationErrorCode
	}

	if strings.Contains(s, AuthorizationErrorString) {
		return AuthorizationErrorCode
	}

	return UnknownErrorCode
}

func RpcErrorHandler(err error) error {
	e := ParseError(err.Error())

	switch e {
	case ValidationErrorCode:
		return status.Errorf(codes.InvalidArgument, "%s", err.Error())
	case NotFoundErrorCode:
		return status.Errorf(codes.NotFound, "%s", err.Error())
	case AuthenticationErrorCode:
		return status.Errorf(codes.Unauthenticated, "%s", err.Error())
	case AuthorizationErrorCode:
		return status.Errorf(codes.PermissionDenied, "%s", err.Error())
	default:
		return status.Errorf(codes.Unknown, "%s", err.Error())
	}
}
