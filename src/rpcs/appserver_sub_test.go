package rpcs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
)

// ----- RPC AppserverSub -----

// Test GetUserAppserverSubs
func TestGetUserAppserverSubs(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.GetUserAppserverSubs(
			ctx, &pb_appserver.GetUserAppserverSubsRequest{},
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
		uId, _ := uuid.Parse(userId)
		appuser := testAppuser(t, &qx.Appuser{ID: uId, Username: "foo"})
		appserver := testAppserver(t, userId, nil)
		appserver2 := testAppserver(t, userId, nil)
		testAppserverSub(t, appuser, appserver)
		testAppserverSub(t, appuser, appserver2)

		// ACT
		response, err := TestAppserverClient.GetUserAppserverSubs(ctx, &pb_appserver.GetUserAppserverSubsRequest{})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

}

// Test GetUserAppserverSubs
func TestGetAllUsersAppserverSubs(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.GetAllUsersAppserverSubs(
			ctx, &pb_appserver.GetAllUsersAppserverSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppusers()))
	})

	t.Run("can_return_all_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		appuser1 := testAppuser(t, nil)
		appuser2 := testAppuser(t, nil)
		appserver := testAppserver(t, userId, nil)
		testAppserverSub(t, appuser1, appserver)
		testAppserverSub(t, appuser2, appserver)

		// ACT
		response, err := TestAppserverClient.GetAllUsersAppserverSubs(ctx, &pb_appserver.GetAllUsersAppserverSubsRequest{
			AppserverId: appserver.ID.String(),
		})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppusers()))
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
		response, err := TestAppserverClient.CreateAppserverSub(ctx, &pb_appserver.CreateAppserverSubRequest{
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
		response, err := TestAppserverClient.CreateAppserverSub(ctx, &pb_appserver.CreateAppserverSubRequest{})
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
		appuser := testAppuser(t, nil)
		appserverSub := testAppserverSub(t, appuser, nil)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverSub(
			ctx, &pb_appserver.DeleteAppserverSubRequest{Id: appserverSub.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.DeleteAppserverSub(ctx, &pb_appserver.DeleteAppserverSubRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
