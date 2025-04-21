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

func TestCreateAppserveRoleSub(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appuser := testAppuser(t, nil)
		appserver := testAppserver(t, nil)
		role := testAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})

		// ACT
		response, err := TestAppserverClient.CreateAppserverRoleSub(
			ctx,
			&pb_appserver.CreateAppserverRoleSubRequest{
				AppserverSubId:  sub.ID.String(),
				AppserverRoleId: role.ID.String(),
				AppserverId:     appserver.ID.String(),
				AppuserId:       appuser.ID.String(),
			},
		)
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

func TestGetAllAppserverUserRoleSubs(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserver := testAppserver(t, nil)

		// ACT
		response, err := TestAppserverClient.GetAllAppserverUserRoleSubs(
			ctx, &pb_appserver.GetAllAppserverUserRoleSubsRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoleSubs()))
	})

	t.Run("can_return_all_appserver_user_sub_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE

		ctx := setup(t, func() {})
		userId, _ := uuid.NewUUID()
		user1 := testAppuser(t, &qx.Appuser{ID: userId, Username: "boo"})
		userId, _ = uuid.NewUUID()
		user2 := testAppuser(t, &qx.Appuser{ID: userId, Username: "bar"})
		appserver := testAppserver(t, nil)
		role := testAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: user1.ID})
		sub2 := testAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: user2.ID})
		testAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: user1.ID, AppserverSubID: sub.ID, AppserverID: appserver.ID,
			},
		)
		testAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: user2.ID, AppserverSubID: sub2.ID, AppserverID: appserver.ID,
			},
		)

		// ACT
		response, err := TestAppserverClient.GetAllAppserverUserRoleSubs(
			ctx, &pb_appserver.GetAllAppserverUserRoleSubsRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoleSubs()))
	})
}

func TestDeleteAppserveRolesSub(t *testing.T) {
	t.Run("roles_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(ctxUserKey).(string))
		appuser := testAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testAppserver(t, &qx.Appserver{Name: "foo", AppuserID: appuser.ID})
		role := testAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})
		roleSub := testAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: appuser.ID, AppserverSubID: sub.ID, AppserverID: appserver.ID,
			},
		)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRoleSub(
			ctx,
			&pb_appserver.DeleteAppserverRoleSubRequest{Id: roleSub.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		asrSub := testAppserverRoleSub(t, nil)

		// ACT
		response, err := TestAppserverClient.DeleteAppserverRoleSub(
			ctx,
			&pb_appserver.DeleteAppserverRoleSubRequest{Id: asrSub.ID.String()},
		)

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
