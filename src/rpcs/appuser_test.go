package rpcs_test

import (
	pb_appuser "mist/src/protos/v1/appuser"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----- RPC CreateAppuser -----
func TestCreateAppuser(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		var count int
		ctx := setup(t, func() {})

		// ACT

		response, err := TestAppuserClient.CreateAppuser(
			ctx,
			&pb_appuser.CreateAppuserRequest{Username: "someone", Id: uuid.NewString()})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		dbConn.QueryRow(ctx, "SELECT COUNT(*) FROM appuser").Scan(&count)
		assert.NotNil(t, response)
		assert.Equal(t, 1, count)
	})

	t.Run("invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppuserClient.CreateAppuser(ctx, &pb_appuser.CreateAppuserRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "missing name attribute")
	})
}
