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
	pb_appserver_permission "mist/src/protos/v1/appserver_permission"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestAppserverPermissionService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, true)
		user := testutil.TestAppuser(t, nil, true)

		// ACT
		response, err := testutil.TestAppserverPermissionClient.Create(ctx, &pb_appserver_permission.CreateRequest{
			AppserverId: appserver.ID.String(),
			AppuserId:   user.ID.String(),
		})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverPermissionClient.Create(ctx, &pb_appserver_permission.CreateRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionWrite, permission.SubActionCreate).Return(
			message.UnauthorizedError("Unauthorized"),
		)
		svc := &rpcs.AppserverPermissionGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&pb_appserver_permission.CreateRequest{AppserverId: uuid.NewString(), AppuserId: uuid.NewString()},
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
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverPermission", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionWrite, permission.SubActionCreate).Return(nil)
		svc := &rpcs.AppserverPermissionGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&pb_appserver_permission.CreateRequest{AppserverId: mockId, AppuserId: mockId},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
	})
}

func TestAppserverPermissionService_ListAppserverUsers(t *testing.T) {
	t.Run("Successful:can_return_empty_list_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, true)

		// ACT
		response, err := testutil.TestAppserverPermissionClient.ListAppserverUsers(ctx, &pb_appserver_permission.ListAppserverUsersRequest{
			AppserverId: appserver.ID.String(),
		})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(response.AppserverPermissions))
	})

	t.Run("Successful:can_return_multiple_users_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, true)
		user1 := testutil.TestAppuser(t, nil, false)
		user2 := testutil.TestAppuser(t, nil, false)
		testutil.TestAppserverPermission(t, &qx.AppserverPermission{AppserverID: appserver.ID, AppuserID: user1.ID}, false)
		testutil.TestAppserverPermission(t, &qx.AppserverPermission{AppserverID: appserver.ID, AppuserID: user2.ID}, false)

		// ACT
		response, err := testutil.TestAppserverPermissionClient.ListAppserverUsers(ctx, &pb_appserver_permission.ListAppserverUsersRequest{
			AppserverId: appserver.ID.String(),
		})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(response.AppserverPermissions))
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := testutil.TestAppserver(t, nil, true).ID
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverPermissions", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead, permission.SubActionListAppserverUserPermsission).Return(nil)
		svc := &rpcs.AppserverPermissionGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		response, err := svc.ListAppserverUsers(ctx, &pb_appserver_permission.ListAppserverUsersRequest{
			AppserverId: appserverId.String(),
		})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "db error")
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nullString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionRead, permission.SubActionListAppserverUserPermsission).Return(
			message.UnauthorizedError("Unauthorized"),
		)
		svc := &rpcs.AppserverPermissionGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListAppserverUsers(
			ctx,
			&pb_appserver_permission.ListAppserverUsersRequest{AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}

func TestAppserverPermissionService_Delete(t *testing.T) {
	t.Run("Successful:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		permissionEntry := testutil.TestAppserverPermission(t, nil, true)

		// ACT
		response, err := testutil.TestAppserverPermissionClient.Delete(ctx, &pb_appserver_permission.DeleteRequest{
			Id: permissionEntry.ID.String(),
		})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverPermissionClient.Delete(ctx, &pb_appserver_permission.DeleteRequest{
			Id: uuid.NewString(),
		})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		permId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &permId, permission.ActionDelete, permission.SubActionDelete).Return(
			message.UnauthorizedError("Unauthorized"),
		)
		svc := &rpcs.AppserverPermissionGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_appserver_permission.DeleteRequest{Id: permId},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverPermission", ctx, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, &mockId, permission.ActionDelete, permission.SubActionDelete).Return(nil)
		svc := &rpcs.AppserverPermissionGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_appserver_permission.DeleteRequest{Id: mockId},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "(-3) database error: db error")
	})
}
