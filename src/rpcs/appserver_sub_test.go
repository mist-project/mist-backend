package rpcs

import (
	"testing"

	pb_server "mist/src/protos/v1/server"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----- RPC AppserverSub -----

// Test GetUserAppserverSubs
func TestGetUserAppserverSubs(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.GetUserAppserverSubs(
			ctx, &pb_server.GetUserAppserverSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("can_return_all_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		testAppserverSub(t, userId, nil)
		testAppserverSub(t, userId, nil)

		// ACT
		response, err := TestAppserverClient.GetUserAppserverSubs(ctx, &pb_server.GetUserAppserverSubsRequest{})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

	t.Run("can_filter_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		testAppserverSub(t, userId, nil)
		testAppserverSub(t, uuid.NewString(), nil)

		// ACT
		response, err := TestAppserverClient.GetUserAppserverSubs(
			ctx, &pb_server.GetUserAppserverSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetAppservers()))
	})
}

// ----- RPC CreateAppserverSub -----
func TestCreateAppserverSub(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		appserver := testAppserver(t, userId, nil)

		// ACT
		response, err := TestAppserverClient.CreateAppserverSub(ctx, &pb_server.CreateAppserverSubRequest{
			AppserverId: appserver.ID.String(),
		})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverSub)
	})

	t.Run("invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.CreateAppserverSub(ctx, &pb_server.CreateAppserverSubRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "(-1): missing appserver_id attribute")
	})
}

// ----- RPC DeleteAllAppserverSubs -----
func TestDeleteAppserverSubs(t *testing.T) {
	t.Run("deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		appserverSub := testAppserverSub(t, userId, nil)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverSub(ctx, &pb_server.DeleteAppserverSubRequest{Id: appserverSub.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.DeleteAppserverSub(ctx, &pb_server.DeleteAppserverSubRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
