package rpcs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mist/src/errors/message"
	"mist/src/permission"
	pb_channel_role "mist/src/protos/v1/channel_role"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestChannelRoleService_ListChannelRoles(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, true)
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: sub.AppserverID},
		)

		// ACT
		response, err := testutil.TestChannelRoleClient.ListChannelRoles(
			ctx, &pb_channel_role.ListChannelRolesRequest{AppserverId: sub.AppserverID.String(), ChannelId: uuid.NewString()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetChannelRoles()))
	})

	t.Run("Successful:can_return_all_appserver_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user := testutil.TestAppuser(t, nil, true)
		role := testutil.TestChannelRole(t, nil, true)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: role.AppserverID, AppuserID: user.ID}, false)

		// ACT
		response, err := testutil.TestChannelRoleClient.ListChannelRoles(
			ctx,
			&pb_channel_role.ListChannelRolesRequest{AppserverId: role.AppserverID.String(), ChannelId: role.ChannelID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetChannelRoles()))
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := testutil.TestAppserver(t, nil, true).ID
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: appserverId},
		)
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListChannelRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead, permission.SubActionListChannelRoles).Return(
			nil,
		)

		svc := &rpcs.ChannelRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		response, err := svc.ListChannelRoles(ctx, &pb_channel_role.ListChannelRolesRequest{
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
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionRead, "list-channel-roles").Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.ChannelRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListChannelRoles(
			ctx,
			&pb_channel_role.ListChannelRolesRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})

}

func TestChannelRoleService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		role := testutil.TestAppserverRole(t, nil, true)
		channel := testutil.TestChannel(t, &qx.Channel{AppserverID: role.AppserverID, Name: "foo"}, true)

		// ACT
		response, err := testutil.TestChannelRoleClient.Create(ctx, &pb_channel_role.CreateRequest{
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
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelRoleClient.Create(ctx, &pb_channel_role.CreateRequest{})
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
		mockAuth.On("Authorize", mock.Anything, nullString, permission.ActionWrite, permission.SubActionCreate).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.ChannelRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&pb_channel_role.CreateRequest{
				ChannelId: uuid.NewString(), AppserverId: uuid.NewString(), AppserverRoleId: uuid.NewString(),
			},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannelRole", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionWrite, permission.SubActionCreate).Return(nil)

		svc := &rpcs.ChannelRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&pb_channel_role.CreateRequest{
				ChannelId: uuid.NewString(), AppserverId: uuid.NewString(), AppserverRoleId: uuid.NewString(),
			},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
	})
}

func TestChannelRoleService_Delete(t *testing.T) {
	t.Run("Successful:roles_can_only_be_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aRole := testutil.TestChannelRole(t, nil, true)

		// ACT
		response, err := testutil.TestChannelRoleClient.Delete(
			ctx, &pb_channel_role.DeleteRequest{Id: aRole.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aRole := testutil.TestChannelRole(t, nil, false)

		// ACT
		response, err := testutil.TestChannelRoleClient.Delete(
			ctx, &pb_channel_role.DeleteRequest{Id: aRole.ID.String()})

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelRoleClient.Delete(ctx, &pb_channel_role.DeleteRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
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

		svc := &rpcs.ChannelRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_channel_role.DeleteRequest{Id: roleId},
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
		mockQuerier.On("DeleteChannelRole", ctx, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, &mockId, permission.ActionDelete, permission.SubActionDelete).Return(
			nil,
		)

		svc := &rpcs.ChannelRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_channel_role.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
	})
}
