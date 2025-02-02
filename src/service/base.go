package service

import (
	"fmt"
	"strings"
)

type CustomGRPCError int

const (
	ValidationError CustomGRPCError = -1
	NotFoundError                   = -2
	DatabaseError                   = -3
	UnknownError                    = -4
)

var (
	DatabaseErrorString   string = fmt.Sprintf("(%d):", DatabaseError)
	NotFoundErrorString   string = fmt.Sprintf("(%d):", NotFoundError)
	validationErrorString string = fmt.Sprintf("(%d):", ValidationError)
)

func ParseServiceError(serviceErr string) CustomGRPCError {

	if strings.Contains(serviceErr, validationErrorString) {
		return ValidationError
	}

	if strings.Contains(serviceErr, DatabaseErrorString) {
		return DatabaseError
	}

	if strings.Contains(serviceErr, NotFoundErrorString) {
		return NotFoundError
	}

	return UnknownError
}

func AddValidationError(attribute string, validationErr []string) []string {
	return append(validationErr, fmt.Sprintf("missing %s attribute", attribute))
}
