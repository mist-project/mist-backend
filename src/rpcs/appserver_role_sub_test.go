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

func TestAppserveRoleSubService_Create(t *testing.T) {
	t.Run("Successfulcreates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appuser := testutil.TestAppuser(t, nil)
		appserver := testutil.TestAppserver(t, nil)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID})
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.Create(
			ctx,
			&pb_appserverrolesub.CreateRequest{
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

	t.Run("Error:on_database_failure_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.Create(
			ctx, &pb_appserverrolesub.CreateRequest{
				AppserverRoleId: uuid.NewString(),
				AppserverSubId:  uuid.NewString(),
				AppserverId:     uuid.NewString(),
				AppuserId:       uuid.NewString(),
			},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.Create(
			ctx, &pb_appserverrolesub.CreateRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestAppserveRoleSubService_ListServerRoleSubs(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.ListServerRoleSubs(
			ctx, &pb_appserverrolesub.ListServerRoleSubsRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoleSubs()))
	})

	t.Run("Successful:can_return_all_appserver_user_sub_roles_for_appserver_successfully", func(t *testing.T) {
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
		response, err := testutil.TestAppserverRoleSubClient.ListServerRoleSubs(
			ctx, &pb_appserverrolesub.ListServerRoleSubsRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoleSubs()))
	})
}

func TestAppserveRoleSubService_Delete(t *testing.T) {
	t.Run("Successful:role_sub_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
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
		response, err := testutil.TestAppserverRoleSubClient.Delete(
			ctx,
			&pb_appserverrolesub.DeleteRequest{Id: roleSub.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		asrSub := testutil.TestAppserverRoleSub(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.Delete(
			ctx,
			&pb_appserverrolesub.DeleteRequest{Id: asrSub.ID.String()},
		)

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-2) resource not found")
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.Delete(
			ctx,
			&pb_appserverrolesub.DeleteRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})
}
