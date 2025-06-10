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
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestChannelRoleRPCService_ListChannelRoles(t *testing.T) {
	t.Run("Success:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		ch := factory.NewFactory(ctx, db).Channel(t, 0, nil)

		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: ch.AppserverID},
		)

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListChannelRoles(
			ctx, &channel_role.ListChannelRolesRequest{AppserverId: ch.AppserverID.String(), ChannelId: ch.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetChannelRoles()))
	})

	t.Run("Success:can_return_all_channel_roles_for_channel_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		f := factory.NewFactory(ctx, db)
		channel := f.Channel(t, 0, nil)
		role := f.AppserverRole(t, 0, nil)
		role1 := f.AppserverRole(t, 1, &qx.AppserverRole{Name: "role1", AppserverID: channel.AppserverID})

		f.ChannelRole(t, 0, &qx.ChannelRole{
			ChannelID:       channel.ID,
			AppserverID:     channel.AppserverID,
			AppserverRoleID: role.ID,
		})

		f.ChannelRole(t, 1, &qx.ChannelRole{
			ChannelID:       channel.ID,
			AppserverID:     channel.AppserverID,
			AppserverRoleID: role1.ID,
		})

		f.ChannelRole(t, 2, nil)

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListChannelRoles(
			ctx,
			&channel_role.ListChannelRolesRequest{
				AppserverId: role.AppserverID.String(), ChannelId: channel.ID.String(),
			},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetChannelRoles()))
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		serverID := uuid.New()
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverID},
		)

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListChannelRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListChannelRoles(ctx, &channel_role.ListChannelRolesRequest{
			AppserverId: serverID.String(),
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

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.ListChannelRoles(
			ctx,
			&channel_role.ListChannelRolesRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})

}

func TestChannelRoleRPCService_Create(t *testing.T) {
	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		_ = factory.UserAppserverOwner(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		channel := f.Channel(t, 0, nil)
		role := f.AppserverRole(t, 0, nil)

		svc := &rpcs.ChannelRoleGRPCService{
			Deps: &rpcs.GrpcDependencies{Db: db, MProducer: testutil.MockRedisProducer}, Auth: testutil.TestMockAuth,
		}

		// ACT
		response, err := svc.Create(ctx, &channel_role.CreateRequest{
			AppserverId:     role.AppserverID.String(),
			AppserverRoleId: role.ID.String(),
			ChannelId:       channel.ID.String(),
		})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.ChannelRole)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelRoleClient.Create(ctx, &channel_role.CreateRequest{})
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

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&channel_role.CreateRequest{
				ChannelId: uuid.NewString(), AppserverId: uuid.NewString(), AppserverRoleId: uuid.NewString(),
			},
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
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&channel_role.CreateRequest{
				ChannelId: uuid.NewString(), AppserverId: uuid.NewString(), AppserverRoleId: uuid.NewString(),
			},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
	})
}

func TestChannelRoleRPCService_Delete(t *testing.T) {
	t.Run("Success:roles_can_be_deleted", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		channelRole := testutil.TestChannelRole(t, nil, true)

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
			ctx, &channel_role.DeleteRequest{Id: channelRole.ID.String(), AppserverId: channelRole.AppserverID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		tu := factory.UserAppserverOwner(t, ctx, db)

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
			ctx, &channel_role.DeleteRequest{Id: uuid.NewString(), AppserverId: tu.Server.ID.String()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), faults.NotFoundMessage)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		roleId := uuid.NewString()
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&channel_role.DeleteRequest{Id: roleId},
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
		mockQuerier.On("GetChannelRoleById", mock.Anything, mock.Anything).Return(qx.ChannelRole{}, nil)
		mockQuerier.On("DeleteChannelRole", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.ChannelRoleGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&channel_role.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})
}
