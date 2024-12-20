package service

import (
  "fmt"
  "strings"
)

type CustomGRPCError int

const (
  ValidationError CustomGRPCError = iota - 1
  DatabaseError
  UnknownError
)

func ParseServiceError(service_error string) CustomGRPCError {
  if strings.Contains(service_error, fmt.Sprintf("(%d):", ValidationError)) {
    return ValidationError
  }

  if strings.Contains(service_error, fmt.Sprintf("(%d):", DatabaseError)) {
    return DatabaseError
  }

  return UnknownError
}
