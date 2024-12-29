package rpcs

import (
	"testing"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/psql_db/qx"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----- RPC AppserveRole -----

// Test GetUserAppserveRoles
func TestGetAllAppserverRoles(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		appserver := testAppserver(t, userId, nil)

		// ACT
		response, err := TestClient.GetAllAppserverRoles(
			ctx, &pb_mistbe.GetAllAppserverRolesRequest{AppserverId: appserver.ID.String()},
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
		userId := ctx.Value(ctxUserKey).(string)
		asRole1 := testAppserverRole(t, userId, nil)
		testAppserverRole(
			t, userId, &qx.AppserverRole{Name: "some random name", AppserverID: asRole1.AppserverID},
		)
		testAppserverRole(t, userId, nil)

		// ACT
		response, err := TestClient.GetAllAppserverRoles(
			ctx, &pb_mistbe.GetAllAppserverRolesRequest{AppserverId: asRole1.AppserverID.String()},
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
		userId := ctx.Value(ctxUserKey).(string)
		appserver := testAppserver(t, userId, nil)

		// ACT
		response, err := TestClient.CreateAppserverRole(ctx, &pb_mistbe.CreateAppserverRoleRequest{
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
		response, err := TestClient.CreateAppserverRole(ctx, &pb_mistbe.CreateAppserverRoleRequest{})
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
		userId := ctx.Value(ctxUserKey).(string)
		appserverRole := testAppserverRole(t, userId, nil)

		// ACT
		response, err := TestClient.DeleteAppserverRole(ctx, &pb_mistbe.DeleteAppserverRoleRequest{Id: appserverRole.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserverRole := testAppserverRole(t, uuid.NewString(), nil)

		// ACT
		response, err := TestClient.DeleteAppserverRole(ctx, &pb_mistbe.DeleteAppserverRoleRequest{Id: appserverRole.ID.String()})

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-2): no rows were deleted")
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.DeleteAppserverRole(ctx, &pb_mistbe.DeleteAppserverRoleRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
