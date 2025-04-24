package rpcs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

// ----- RPC AppserveRole -----

func TestGetAllAppserverRoles(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleClient.GetAllAppserverRoles(
			ctx, &pb_appserverrole.GetAllAppserverRolesRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoles()))
	})

	t.Run("can_return_all_appserver_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil)
		testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "some random name", AppserverID: server.ID})
		testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "some random name #2", AppserverID: server.ID})

		// ACT
		response, err := testutil.TestAppserverRoleClient.GetAllAppserverRoles(
			ctx, &pb_appserverrole.GetAllAppserverRolesRequest{AppserverId: server.ID.String()},
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
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleClient.CreateAppserverRole(ctx, &pb_appserverrole.CreateAppserverRoleRequest{
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
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.CreateAppserverRole(ctx, &pb_appserverrole.CreateAppserverRoleRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

// ----- RPC DeleteAllAppserveRoles -----
func TestDeleteAppserveRoles(t *testing.T) {
	t.Run("roles_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		user := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		server := testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: user.ID})
		aRole := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: server.ID, Name: "zoo"})

		// ACT
		response, err := testutil.TestAppserverRoleClient.DeleteAppserverRole(
			ctx, &pb_appserverrole.DeleteAppserverRoleRequest{Id: aRole.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aRole := testutil.TestAppserverRole(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleClient.DeleteAppserverRole(
			ctx, &pb_appserverrole.DeleteAppserverRoleRequest{Id: aRole.ID.String()})

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-2): no rows were deleted")
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.DeleteAppserverRole(ctx, &pb_appserverrole.DeleteAppserverRoleRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
