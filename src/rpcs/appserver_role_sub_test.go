package rpcs_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mist/src/errors/message"
	"mist/src/permission"
	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestAppserveRoleSubService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appuser := testutil.TestAppuser(t, nil, false)
		appserver := testutil.TestAppserver(t, nil, true)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID}, false)
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID}, false)

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
		sub := testutil.TestAppserverSub(t, nil, true)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.ListServerRoleSubs(
			ctx, &pb_appserverrolesub.ListServerRoleSubsRequest{AppserverId: sub.AppserverID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoleSubs()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionRead, permission.SubActionListAppserverUserRoleSubs).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListServerRoleSubs(
			ctx,
			&pb_appserverrolesub.ListServerRoleSubsRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})

	t.Run("Successful:can_return_all_appserver_user_sub_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		userId, _ := uuid.NewUUID()
		user1 := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "boo"}, false)
		userId, _ = uuid.NewUUID()
		user2 := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "bar"}, false)
		mainSub := testutil.TestAppserverSub(t, nil, true)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: mainSub.AppserverID}, false)
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: mainSub.AppserverID, AppuserID: user1.ID}, false)
		sub2 := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: mainSub.AppserverID, AppuserID: user2.ID}, false)
		testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: user1.ID, AppserverSubID: sub.ID, AppserverID: sub.AppserverID,
			},
			false,
		)
		testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: user2.ID, AppserverSubID: sub2.ID, AppserverID: sub.AppserverID,
			},
			false,
		)

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.ListServerRoleSubs(
			ctx, &pb_appserverrolesub.ListServerRoleSubsRequest{AppserverId: sub.AppserverID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoleSubs()))
	})
}

func TestAppserveRoleSubService_Delete(t *testing.T) {
	t.Run("Successful:role_sub_can_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appuser := testutil.TestAppuser(t, nil, false)
		appserver := testutil.TestAppserver(t, nil, true)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: appserver.ID}, false)
		sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID}, false)
		roleSub := testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppserverRoleID: role.ID, AppuserID: appuser.ID, AppserverSubID: sub.ID, AppserverID: appserver.ID,
			},
			false,
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

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		roleId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete, permission.SubActionDelete).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_appserverrolesub.DeleteRequest{Id: roleId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRoleSub", ctx, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, &mockId, permission.ActionDelete, permission.SubActionDelete).Return(
			nil,
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_appserverrolesub.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
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
