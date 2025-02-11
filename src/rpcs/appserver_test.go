package rpcs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

// ----- RPC Appservers -----
func TestListAppServer(t *testing.T) {
	t.Run("can_returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.ListAppservers(
			ctx, &pb_appserver.ListAppserversRequest{Name: wrapperspb.String("random")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("can_return_all_resources_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)

		testAppserver(t, userId, nil)
		testAppserver(t, userId, &qx.Appserver{Name: "another one"})

		// ACT
		response, err := TestAppserverClient.ListAppservers(ctx, &pb_appserver.ListAppserversRequest{})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

	t.Run("can_filter_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		testAppserver(t, uuid.NewString(), nil)
		testAppserver(t, userId, &qx.Appserver{Name: "another one"})

		// ACT
		response, err := TestAppserverClient.ListAppservers(
			ctx, &pb_appserver.ListAppserversRequest{Name: wrapperspb.String("another one")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetAppservers()))
	})
}

// ----- RPC GetByIdAppserver -----

func TestGetByIdAppServer(t *testing.T) {
	t.Run("returns_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		appserver := testAppserver(t, userId, nil)

		// ACT
		response, err := TestAppserverClient.GetByIdAppserver(
			ctx, &pb_appserver.GetByIdAppserverRequest{Id: appserver.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, appserver.ID.String(), response.GetAppserver().Id)
		assert.Equal(t, true, response.GetAppserver().IsOwner)
		assert.Equal(t, appserver.Name, response.GetAppserver().Name)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.GetByIdAppserver(
			ctx, &pb_appserver.GetByIdAppserverRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})

	t.Run("invalid_uuid_returns_parsing_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.GetByIdAppserver(
			ctx, &pb_appserver.GetByIdAppserverRequest{Id: "foo"},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "invalid UUID")
	})
}

// ----- RPC CreateAppserver -----
func TestCreateAppserver(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		var count int
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		parsedUserId, err := uuid.Parse(userId)
		testAppuser(t, &qx.Appuser{ID: parsedUserId, Username: "foo"})

		// ACT
		response, err := TestAppserverClient.CreateAppserver(ctx, &pb_appserver.CreateAppserverRequest{Name: "someone"})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		dbcPool.QueryRow(ctx, "SELECT COUNT(*) FROM appserver").Scan(&count)

		serverSubs, _ := service.NewAppserverSubService(dbcPool, ctx).ListUserAppserverAndSub(userId)
		assert.NotNil(t, response.Appserver)
		assert.Equal(t, 1, len(serverSubs))
		assert.Equal(t, 1, count)
	})

	t.Run("invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.CreateAppserver(ctx, &pb_appserver.CreateAppserverRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "missing name attribute")
	})
}

// ----- RPC Deleteappserver -----
func TestDeleteAppserver(t *testing.T) {

	t.Run("deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		appserver := testAppserver(t, userId, nil)
		testAppserverSub(t, userId, &qx.AppserverSub{AppserverID: appserver.ID})
		ass := service.NewAppserverSubService(dbcPool, ctx)

		// ASSERT
		serverSubs, _ := ass.ListUserAppserverAndSub(userId)
		assert.Equal(t, 1, len(serverSubs))

		// ACT
		response, err := TestAppserverClient.DeleteAppserver(ctx, &pb_appserver.DeleteAppserverRequest{Id: appserver.ID.String()})

		// ASSERT
		serverSubs, _ = ass.ListUserAppserverAndSub(userId)
		assert.NotNil(t, response)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(serverSubs))
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.DeleteAppserver(ctx, &pb_appserver.DeleteAppserverRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
