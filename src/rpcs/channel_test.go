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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/permission"
	"mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestChannelRPCService_ListServerChannels(t *testing.T) {
	t.Run("Successful:returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, true)
		ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: sub.AppserverID})

		// ACT
		response, err := testutil.TestChannelClient.ListServerChannels(
			ctx, &channel.ListServerChannelsRequest{AppserverId: sub.AppserverID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetChannels()))
	})

	t.Run("Successful:returns_all_resources_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		subId := testutil.TestAppserverSub(t, nil, true)
		serverId := subId.AppserverID
		testutil.TestChannel(t, nil, false)
		testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: serverId}, false)
		testutil.TestChannel(t, &qx.Channel{Name: "bar", AppserverID: serverId}, false)
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
		)

		// ACT
		response, err := testutil.TestChannelClient.ListServerChannels(ctx, &channel.ListServerChannelsRequest{
			AppserverId: serverId.String(),
		})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetChannels()))
	})

	t.Run("Successful:can_filter_by_name", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		subId := testutil.TestAppserverSub(t, nil, true)
		serverId := subId.AppserverID
		testutil.TestChannel(t, nil, false)
		testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: serverId}, false)
		testutil.TestChannel(t, &qx.Channel{Name: "bar", AppserverID: serverId}, false)
		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
		)

		// ACT
		response, err := testutil.TestChannelClient.ListServerChannels(ctx, &channel.ListServerChannelsRequest{
			AppserverId: serverId.String(), Name: wrapperspb.String("foo"),
		})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetChannels()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		serverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListServerChannels(
			ctx,
			&channel.ListServerChannelsRequest{AppserverId: serverId.String()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
	})
}

func TestChannelRPCService_GetById(t *testing.T) {
	t.Run("Successful:returns_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, true)
		c := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: sub.AppserverID}, true)

		// ACT
		response, err := testutil.TestChannelClient.GetById(
			ctx, &channel.GetByIdRequest{Id: c.ID.String(), AppserverId: sub.AppserverID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, c.ID.String(), response.GetChannel().Id)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		tu := factory.UserAppserverWithAllPermissions(t)

		// ACT
		response, err := testutil.TestChannelClient.GetById(
			ctx, &channel.GetByIdRequest{Id: uuid.NewString(), AppserverId: tu.Server.AppuserID.String()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), faults.NotFoundMessage)
	})

	t.Run("Error:when_get_by_id_search_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		channelId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &channelId, permission.ActionRead).Return(
			nil,
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.GetById(
			ctx,
			&channel.GetByIdRequest{Id: channelId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
	})

	t.Run("Error:invalid_uuid_returns_parsing_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.GetById(
			ctx, &channel.GetByIdRequest{Id: "foo"},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error:\n - id: value must be a valid UUID")
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		mockId := uuid.NewString()
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &mockId, permission.ActionRead).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.GetById(
			ctx,
			&channel.GetByIdRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
	})
}

func TestChannelRPCService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, true)

		// ACT
		response, err := testutil.TestChannelClient.Create(
			ctx,
			&channel.CreateRequest{Name: "new channel", AppserverId: sub.AppserverID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.Channel)
	})

	t.Run("Error:when_create_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannel", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionCreate).Return(
			nil,
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&channel.CreateRequest{Name: "foo", AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-3) create channel error: db error")
	})

	t.Run("Error:invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.Create(ctx, &channel.CreateRequest{})
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
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionCreate).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&channel.CreateRequest{Name: "foo", AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
	})
}

func TestChannelRPCService_Delete(t *testing.T) {
	t.Run("Successful:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		c := testutil.TestChannel(t, nil, true)

		// ACT
		response, err := testutil.TestChannelClient.Delete(
			ctx, &channel.DeleteRequest{Id: c.ID.String(), AppserverId: c.AppserverID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		c := testutil.TestChannel(t, nil, true)

		// ACT
		response, err := testutil.TestChannelClient.Delete(
			ctx, &channel.DeleteRequest{Id: uuid.NewString(), AppserverId: c.AppserverID.String()},
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
		mockId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &mockId, permission.ActionDelete).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&channel.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{ID: uuid.New()}, nil)
		mockQuerier.On("DeleteChannel", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &mockId, permission.ActionDelete).Return(
			nil,
		)

		svc := &rpcs.ChannelGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&channel.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
	})
}
