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
		parsedUid, _ := uuid.Parse(ctx.Value(ctxUserKey).(string))
		appuser := testAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testAppserver(t, nil)
		appserver2 := testAppserver(t, nil)
		testAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})
		testAppserverSub(t, &qx.AppserverSub{AppserverID: appserver2.ID, AppuserID: appuser.ID})

		// ACT
		response, err := TestAppserverClient.GetUserAppserverSubs(
			ctx, &pb_appserver.GetUserAppserverSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

}

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
		user1 := testAppuser(t, nil)
		user2 := testAppuser(t, nil)
		server := testAppserver(t, nil)
		testAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user1.ID})
		testAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user2.ID})

		// ACT
		response, err := TestAppserverClient.GetAllUsersAppserverSubs(
			ctx,
			&pb_appserver.GetAllUsersAppserverSubsRequest{AppserverId: server.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppusers()))
	})

}

func TestCreateAppserverSub(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(ctxUserKey).(string))
		appuser := testAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testAppserver(t, &qx.Appserver{AppuserID: appuser.ID})

		// ACT
		response, err := TestAppserverClient.CreateAppserverSub(
			ctx, &pb_appserver.CreateAppserverSubRequest{AppserverId: appserver.ID.String()},
		)

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
		response, err := TestAppserverClient.CreateAppserverSub(
			ctx, &pb_appserver.CreateAppserverSubRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "(-1): missing appserver_id attribute")
	})
}

func TestDeleteAppserverSubs(t *testing.T) {
	t.Run("deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(ctxUserKey).(string))
		appuser := testAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testAppserver(t, &qx.Appserver{AppuserID: appuser.ID})
		appserverSub := testAppserverSub(t, &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   appuser.ID,
		})

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
		response, err := TestAppserverClient.DeleteAppserverSub(
			ctx, &pb_appserver.DeleteAppserverSubRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
