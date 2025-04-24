package rpcs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

func TestCreateAppserveRoleSub(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appuser := testutil.TestAppuser(t, nil)
		appserver := testutil.TestAppserver(t, nil)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.CreateAppserverRoleSub(
			ctx,
			&pb_appserverrolesub.CreateAppserverRoleSubRequest{
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
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.CreateAppserverRoleSub(
			ctx, &pb_appserverrolesub.CreateAppserverRoleSubRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestGetAllAppserverUserRoleSubs(t *testing.T) {
	t.Run("can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.GetAllAppserverUserRoleSubs(
			ctx, &pb_appserverrolesub.GetAllAppserverUserRoleSubsRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoleSubs()))
	})

	t.Run("can_return_all_appserver_user_sub_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE

		ctx := testutil.Setup(t, func() {})
		userId, _ := uuid.NewUUID()
		user1 := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "boo"})
		userId, _ = uuid.NewUUID()
		user2 := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "bar"})
		appserver := testutil.TestAppserver(t, nil)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: user1.ID})
		sub2 := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: user2.ID})
		testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: user1.ID, AppserverSubID: sub.ID, AppserverID: appserver.ID,
			},
		)
		testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: user2.ID, AppserverSubID: sub2.ID, AppserverID: appserver.ID,
			},
		)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.GetAllAppserverUserRoleSubs(
			ctx, &pb_appserverrolesub.GetAllAppserverUserRoleSubsRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoleSubs()))
	})
}

func TestDeleteAppserveRoleSub(t *testing.T) {
	t.Run("roles_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "delete role sub user"})
		appserver := testutil.TestAppserver(t, &qx.Appserver{Name: "foo", AppuserID: appuser.ID})
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})
		roleSub := testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: appuser.ID, AppserverSubID: sub.ID, AppserverID: appserver.ID,
			},
		)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.DeleteAppserverRoleSub(
			ctx,
			&pb_appserverrolesub.DeleteAppserverRoleSubRequest{Id: roleSub.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		asrSub := testutil.TestAppserverRoleSub(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.DeleteAppserverRoleSub(
			ctx,
			&pb_appserverrolesub.DeleteAppserverRoleSubRequest{Id: asrSub.ID.String()},
		)

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-2): no rows were deleted")
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.DeleteAppserverRoleSub(
			ctx,
			&pb_appserverrolesub.DeleteAppserverRoleSubRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
