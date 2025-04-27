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

func TestAppserverSubService_ListUserServerSubs(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.ListUserServerSubs(
			ctx, &pb_appserversub.ListUserServerSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("Successful:can_return_all_users_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		appserver := testutil.TestAppserver(t, nil, false)
		appserver2 := testutil.TestAppserver(t, nil, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID}, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver2.ID, AppuserID: appuser.ID}, false)

		// ACT
		response, err := testutil.TestAppserverSubClient.ListUserServerSubs(
			ctx, &pb_appserversub.ListUserServerSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

}

func TestAppserverSubService_ListAppserverUserSubs(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.ListAppserverUserSubs(
			ctx, &pb_appserversub.ListAppserverUserSubsRequest{AppserverId: uuid.NewString()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppusers()))
	})

	t.Run("Successful:can_return_all_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user1 := testutil.TestAppuser(t, &qx.Appuser{ID: uuid.New(), Username: "foo"}, false)
		user2 := testutil.TestAppuser(t, &qx.Appuser{ID: uuid.New(), Username: "bar"}, false)
		server := testutil.TestAppserver(t, nil, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user1.ID}, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user2.ID}, false)

		// ACT
		response, err := testutil.TestAppserverSubClient.ListAppserverUserSubs(
			ctx,
			&pb_appserversub.ListAppserverUserSubsRequest{AppserverId: server.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppusers()))
	})

}

func TestAppserverSubService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: appuser.ID}, false)

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(
			ctx, &pb_appserversub.CreateRequest{AppserverId: appserver.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverSub)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(
			ctx, &pb_appserversub.CreateRequest{AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "database error")
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(
			ctx, &pb_appserversub.CreateRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestAppserverSubService_Delete(t *testing.T) {
	t.Run("Successful:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: appuser.ID}, false)
		appserverSub := testutil.TestAppserverSub(t, &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   appuser.ID,
		},
			false,
		)

		// ACT
		response, err := testutil.TestAppserverSubClient.Delete(
			ctx, &pb_appserversub.DeleteRequest{Id: appserverSub.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Delete(
			ctx, &pb_appserversub.DeleteRequest{Id: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})
}
