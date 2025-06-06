package rpcs_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestAppserverRoleRPCService_ListServerRoles(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, true)
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: sub.AppserverID},
		)

		// ACT
		response, err := testutil.TestAppserverRoleClient.ListServerRoles(
			ctx, &appserver_role.ListServerRolesRequest{AppserverId: sub.AppserverID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoles()))
	})

	t.Run("Successful:can_return_all_appserver_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, true)
		testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "some random name", AppserverID: sub.AppserverID}, false)
		testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "some random name #2", AppserverID: sub.AppserverID}, false)

		// ACT
		response, err := testutil.TestAppserverRoleClient.ListServerRoles(
			ctx, &appserver_role.ListServerRolesRequest{AppserverId: sub.AppserverID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoles()))
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := testutil.TestAppserver(t, nil, true).ID
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: appserverId},
		)
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead).Return(
			nil,
		)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		response, err := svc.ListServerRoles(ctx, &appserver_role.ListServerRolesRequest{
			AppserverId: appserverId.String(),
		})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code())
		assert.Contains(t, s.Message(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nullString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListServerRoles(
			ctx,
			&appserver_role.ListServerRolesRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

}

func TestAppserverRoleRPCService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, true)

		// ACT
		response, err := testutil.TestAppserverRoleClient.Create(ctx, &appserver_role.CreateRequest{
			AppserverId:             appserver.ID.String(),
			Name:                    "foo",
			AppserverPermissionMask: 0,
			ChannelPermissionMask:   0,
			SubPermissionMask:       0,
		})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverRole)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.Create(ctx, &appserver_role.CreateRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nullString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionCreate).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&appserver_role.CreateRequest{Name: "foo", AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRole", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionCreate).Return(nil)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&appserver_role.CreateRequest{Name: "foo", AppserverId: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})
}

func TestAppserverRoleRPCService_Delete(t *testing.T) {
	t.Run("Successful:roles_can_only_be_deleted_by_server_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aRole := testutil.TestAppserverRole(t, nil, true)

		// ACT
		response, err := testutil.TestAppserverRoleClient.Delete(
			ctx, &appserver_role.DeleteRequest{Id: aRole.ID.String(), AppserverId: aRole.AppserverID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aRole := testutil.TestAppserverRole(t, nil, false)

		// ACT
		response, err := testutil.TestAppserverRoleClient.Delete(
			ctx, &appserver_role.DeleteRequest{Id: aRole.ID.String(), AppserverId: aRole.AppserverID.String()})

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.Delete(
			ctx, &appserver_role.DeleteRequest{Id: uuid.NewString(), AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, err.Error(), faults.NotFoundMessage)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		roleId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&appserver_role.DeleteRequest{Id: roleId, AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &mockId, permission.ActionDelete).Return(
			nil,
		)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&appserver_role.DeleteRequest{Id: mockId, AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})
}
