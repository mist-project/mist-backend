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

// ----- RPC AppserveRole -----

func TestGetAllAppserverRoles(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserver := testAppserver(t, nil)

		// ACT
		response, err := TestAppserverClient.GetAllAppserverRoles(
			ctx, &pb_appserver.GetAllAppserverRolesRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoles()))
	})

	t.Run("can_return_all_appserver_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		server := testAppserver(t, nil)
		testAppserverRole(t, &qx.AppserverRole{Name: "some random name", AppserverID: server.ID})
		testAppserverRole(t, &qx.AppserverRole{Name: "some random name #2", AppserverID: server.ID})

		// ACT
		response, err := TestAppserverClient.GetAllAppserverRoles(
			ctx, &pb_appserver.GetAllAppserverRolesRequest{AppserverId: server.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoles()))
	})
}

// ----- RPC CreateAppserveRole -----
func TestCreateAppserveRole(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserver := testAppserver(t, nil)

		// ACT
		response, err := TestAppserverClient.CreateAppserverRole(ctx, &pb_appserver.CreateAppserverRoleRequest{
			AppserverId: appserver.ID.String(),
			Name:        "foo",
		})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverRole)
	})

	t.Run("invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.CreateAppserverRole(ctx, &pb_appserver.CreateAppserverRoleRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "(-1): missing appserver_id attribute, missing name attribute")
	})
}

// ----- RPC DeleteAllAppserveRoles -----
func TestDeleteAppserveRoles(t *testing.T) {
	t.Run("roles_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(ctxUserKey).(string))
		user := testAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		server := testAppserver(t, &qx.Appserver{Name: "bar", AppuserID: user.ID})
		aRole := testAppserverRole(t, &qx.AppserverRole{AppserverID: server.ID, Name: "zoo"})

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRole(
			ctx, &pb_appserver.DeleteAppserverRoleRequest{Id: aRole.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		aRole := testAppserverRole(t, nil)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRole(
			ctx, &pb_appserver.DeleteAppserverRoleRequest{Id: aRole.ID.String()})

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-2): no rows were deleted")
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRole(ctx, &pb_appserver.DeleteAppserverRoleRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
