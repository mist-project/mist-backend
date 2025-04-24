package rpcs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_appserversub "mist/src/protos/v1/appserver_sub"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

func TestGetUserAppserverSubs(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.GetUserAppserverSubs(
			ctx, &pb_appserversub.GetUserAppserverSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("can_return_all_users_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testutil.TestAppserver(t, nil)
		appserver2 := testutil.TestAppserver(t, nil)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver2.ID, AppuserID: appuser.ID})

		// ACT
		response, err := testutil.TestAppserverSubClient.GetUserAppserverSubs(
			ctx, &pb_appserversub.GetUserAppserverSubsRequest{},
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
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.GetAllUsersAppserverSubs(
			ctx, &pb_appserversub.GetAllUsersAppserverSubsRequest{AppserverId: uuid.NewString()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppusers()))
	})

	t.Run("can_return_all_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user1 := testutil.TestAppuser(t, &qx.Appuser{ID: uuid.New(), Username: "foo"})
		user2 := testutil.TestAppuser(t, &qx.Appuser{ID: uuid.New(), Username: "bar"})
		server := testutil.TestAppserver(t, nil)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user1.ID})
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user2.ID})

		// ACT
		response, err := testutil.TestAppserverSubClient.GetAllUsersAppserverSubs(
			ctx,
			&pb_appserversub.GetAllUsersAppserverSubsRequest{AppserverId: server.ID.String()},
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
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: appuser.ID})

		// ACT
		response, err := testutil.TestAppserverSubClient.CreateAppserverSub(
			ctx, &pb_appserversub.CreateAppserverSubRequest{AppserverId: appserver.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverSub)
	})

	t.Run("invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.CreateAppserverSub(
			ctx, &pb_appserversub.CreateAppserverSubRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestDeleteAppserverSubs(t *testing.T) {
	t.Run("deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: appuser.ID})
		appserverSub := testutil.TestAppserverSub(t, &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   appuser.ID,
		})

		// ACT
		response, err := testutil.TestAppserverSubClient.DeleteAppserverSub(
			ctx, &pb_appserversub.DeleteAppserverSubRequest{Id: appserverSub.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.DeleteAppserverSub(
			ctx, &pb_appserversub.DeleteAppserverSubRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
