package service_test

import (
	"fmt"
	"testing"

	"mist/src/service"

	"github.com/stretchr/testify/assert"
)

func TestParseServiceError(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput service.CustomGRPCError
	}{
		{
			name:       "should return ValidationError",
			input:      fmt.Sprintf("(%d) validation error", service.ValidationError),
			wantOutput: service.ValidationError,
		},
		{
			name:       "should return DatabaseError",
			input:      fmt.Sprintf("(%d) db error", service.DatabaseError),
			wantOutput: service.DatabaseError,
		},
		{
			name:       "should return NotFoundError",
			input:      fmt.Sprintf("(%d) not found", service.NotFoundError),
			wantOutput: service.NotFoundError,
		},
		{
			name:       "should return UnknownError for unrelated string",
			input:      "some unknown issue occurred",
			wantOutput: service.UnknownError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ACT
			result := service.ParseServiceError(tt.input)

			// ASSERT
			assert.Equal(t, tt.wantOutput, result)
		})
	}
}

func TestAddValidationError(t *testing.T) {
	// ARRANGE
	attribute := "username"
	initialErrors := []string{"missing password attribute"}
	expected := []string{"missing password attribute", "missing username attribute"}

	// ACT
	result := service.AddValidationError(attribute, initialErrors)

	// ASSERT
	assert.Equal(t, expected, result)
}
