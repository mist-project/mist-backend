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

// ----- RPC CreateAppserveRoleSub -----
func TestCreateAppserveRoleSub(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		ownerId, _ := uuid.Parse(userId)
		asRole := testAppserverRole(t, userId, nil)
		asSub := testAppserverSub(t, userId, &qx.AppserverSub{AppserverID: asRole.AppserverID, AppuserID: ownerId})

		// ACT
		response, err := TestAppserverClient.CreateAppserverRoleSub(ctx, &pb_appserver.CreateAppserverRoleSubRequest{
			AppserverSubId:  asSub.ID.String(),
			AppserverRoleId: asRole.ID.String(),
		})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverRoleSub)
	})

	t.Run("invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppserverClient.CreateAppserverRoleSub(ctx, &pb_appserver.CreateAppserverRoleSubRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "(-1): missing appserver_role_id attribute, missing appserver_sub_id attribute")
	})
}

// ----- RPC DeleteAllAppserveRoles -----
func TestDeleteAppserveRolesSub(t *testing.T) {
	t.Run("roles_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		userId := ctx.Value(ctxUserKey).(string)
		asrSub := testAppserverRoleSub(t, userId, nil)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRoleSub(ctx, &pb_appserver.DeleteAppserverRoleSubRequest{Id: asrSub.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		asrSub := testAppserverRoleSub(t, uuid.NewString(), nil)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRoleSub(ctx, &pb_appserver.DeleteAppserverRoleSubRequest{Id: asrSub.ID.String()})

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
