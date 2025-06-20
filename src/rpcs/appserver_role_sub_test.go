package rpcs_test

import (
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
	"mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestAppserveRoleSubRPCService_Create(t *testing.T) {
	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)
		role := factory.NewFactory(ctx, db).AppserverRole(
			t, 0, &qx.AppserverRole{Name: "foo", AppserverID: su.Server.ID},
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Create(
			ctx,
			&appserver_role_sub.CreateRequest{
				AppserverSubId:  su.Sub.ID.String(),
				AppserverRoleId: role.ID.String(),
				AppserverId:     su.Server.ID.String(),
				AppuserId:       su.User.ID.String(),
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
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRoleSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Create(
			ctx, &appserver_role_sub.CreateRequest{
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
		assert.Equal(t, codes.Internal, s.Code())
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var mockId *string
		ctx, db := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mockId, permission.ActionCreate).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		response, err := svc.Create(
			ctx, &appserver_role_sub.CreateRequest{
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
		assert.Equal(t, codes.PermissionDenied, s.Code())
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleSubClient.Create(
			ctx, &appserver_role_sub.CreateRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestAppserveRoleSubRPCService_ListServerRoleSubs(t *testing.T) {
	t.Run("Success:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}
		// Ensure there are no role subs for this sub
		// ACT
		response, err := svc.ListServerRoleSubs(
			ctx, &appserver_role_sub.ListServerRoleSubsRequest{AppserverId: su.Server.ID.String()},
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
		ctx, db := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)

		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.ListServerRoleSubs(
			ctx,
			&appserver_role_sub.ListServerRoleSubsRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})

	t.Run("Success:can_return_all_appserver_user_sub_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		user1 := f.Appuser(t, 2, nil)
		user2 := f.Appuser(t, 3, nil)
		sub1 := f.AppserverSub(t, 2, &qx.AppserverSub{AppserverID: su.Server.ID, AppuserID: user1.ID})
		sub2 := f.AppserverSub(t, 3, &qx.AppserverSub{AppserverID: su.Server.ID, AppuserID: user2.ID})

		role := f.AppserverRole(t, 0, &qx.AppserverRole{Name: "foo", AppserverID: su.Server.ID})
		f.AppserverRoleSub(t, 0, &qx.AppserverRoleSub{
			AppserverRoleID: role.ID,
			AppuserID:       user1.ID,
			AppserverSubID:  sub1.ID,
			AppserverID:     su.Server.ID,
		})
		f.AppserverRoleSub(t, 0, &qx.AppserverRoleSub{
			AppserverRoleID: role.ID,
			AppuserID:       user2.ID,
			AppserverSubID:  sub2.ID,
			AppserverID:     su.Server.ID,
		})

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListServerRoleSubs(
			ctx, &appserver_role_sub.ListServerRoleSubsRequest{AppserverId: su.Server.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoleSubs()))
	})
}

func TestAppserveRoleSubRPCService_Delete(t *testing.T) {
	t.Run("Success:can_successfully_delete_appserver_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		role := f.AppserverRole(t, 0, &qx.AppserverRole{Name: "foo", AppserverID: su.Server.ID})
		roleSub := f.AppserverRoleSub(t, 0, &qx.AppserverRoleSub{
			AppserverRoleID: role.ID,
			AppuserID:       su.User.ID,
			AppserverSubID:  su.Sub.ID,
			AppserverID:     su.Server.ID,
		})

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
			ctx,
			&appserver_role_sub.DeleteRequest{Id: roleSub.ID.String(), AppserverId: su.Server.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		roleId := uuid.NewString()

		mockAuth := new(testutil.MockAuthorizer)

		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&appserver_role_sub.DeleteRequest{Id: roleId},
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

		mockQuerier.On("GetAppserverRoleSubById", mock.Anything, mock.Anything).Return(qx.AppserverRoleSub{}, nil)
		mockQuerier.On("DeleteAppserverRoleSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&appserver_role_sub.DeleteRequest{Id: mockId, AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverRoleSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
			ctx,
			&appserver_role_sub.DeleteRequest{Id: uuid.NewString(), AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), faults.NotFoundMessage)
	})
}
