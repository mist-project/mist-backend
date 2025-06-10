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
	"mist/src/testutil/factory"
)

func TestAppserverRoleRPCService_ListServerRoles(t *testing.T) {
	t.Run("Success:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: su.Server.ID},
		)

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListServerRoles(
			ctx, &appserver_role.ListServerRolesRequest{AppserverId: su.Server.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoles()))
	})

	t.Run("Success:can_return_all_appserver_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		factory.NewFactory(ctx, db).AppserverRole(
			t, 0, &qx.AppserverRole{AppserverID: uuid.MustParse(su.Server.ID.String()), Name: "foo"},
		)
		factory.NewFactory(ctx, db).AppserverRole(
			t, 1, &qx.AppserverRole{AppserverID: uuid.MustParse(su.Server.ID.String()), Name: "bar"},
		)

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListServerRoles(
			ctx, &appserver_role.ListServerRolesRequest{AppserverId: su.Server.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoles()))
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: appserverId},
		)

		mockQuerier := new(testutil.MockQuerier)

		mockQuerier.On("ListAppserverRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

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
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nullString *string
		ctx, db := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

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
		mockAuth.AssertExpectations(t)
	})

}

func TestAppserverRoleRPCService_Create(t *testing.T) {
	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		sub := factory.UserAppserverOwner(t, ctx, db)

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Create(ctx, &appserver_role.CreateRequest{
			AppserverId:             sub.Server.ID.String(),
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
		ctx, _ := testutil.Setup(t, func() {})

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
		ctx, db := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionCreate).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

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
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx, _ := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRole", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

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
	})
}

func TestAppserverRoleRPCService_Delete(t *testing.T) {
	t.Run("Success:roles_can_be_deleted", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)
		role := factory.NewFactory(ctx, db).AppserverRole(t, 0, &qx.AppserverRole{AppserverID: su.Server.ID, Name: "foo"})

		svc := &rpcs.AppserverRoleGRPCService{
			Deps: &rpcs.GrpcDependencies{Db: db, MProducer: testutil.MockRedisProducer}, Auth: testutil.TestMockAuth,
		}

		// ACT
		response, err := svc.Delete(
			ctx, &appserver_role.DeleteRequest{Id: role.ID.String(), AppserverId: role.AppserverID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
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
		ctx, db := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

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
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

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
	})
}
