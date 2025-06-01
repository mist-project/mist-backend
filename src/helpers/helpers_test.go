package helpers_test

import (
	"context"
	"mist/src/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestId(t *testing.T) {
	t.Run("it_returns_the_request_id_from_context", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		expectedRequestId := "test-request-id"
		ctx = context.WithValue(ctx, helpers.RequestIdKey, expectedRequestId)

		// ACT
		requestId := helpers.GetRequestId(ctx)

		// ASSERT
		assert.Equal(t, expectedRequestId, requestId, "Expected to retrieve the request ID from context")
	})

	t.Run("it_returns_a_temporary_request_id_when_not_in_context", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// ACT
		requestId := helpers.GetRequestId(ctx)

		// ASSERT
		assert.NotNil(t, requestId, "Expected a temporary request ID to be generated when not in context")
	})
}
