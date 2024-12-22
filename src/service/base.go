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

var validationErrorString string = fmt.Sprintf("(%d):", ValidationError)
var DatabaseErrorString string = fmt.Sprintf("(%d):", DatabaseError)
var NotFoundErrorString string = fmt.Sprintf("(%d):", NotFoundError)

func ParseServiceError(service_error string) CustomGRPCError {
	if strings.Contains(service_error, validationErrorString) {
		return ValidationError
	}

	if strings.Contains(service_error, DatabaseErrorString) {
		return DatabaseError
	}

	if strings.Contains(service_error, NotFoundErrorString) {
		return NotFoundError
	}

	return UnknownError
}
