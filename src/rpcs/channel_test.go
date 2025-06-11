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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestChannelRPCService_ListServerChannels(t *testing.T) {
	t.Run("Success:returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: su.Server.ID})

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListServerChannels(
			ctx, &channel.ListServerChannelsRequest{AppserverId: su.Server.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetChannels()))
	})

	t.Run("Success:returns_all_resources_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		f := factory.NewFactory(ctx, db)
		s := f.Appserver(t, 0, nil)
		c0 := f.Channel(t, 0, &qx.Channel{Name: "foo", AppserverID: s.ID})
		f.Channel(t, 1, &qx.Channel{Name: "bar", AppserverID: s.ID})
		f.Channel(t, 2, nil)

		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: c0.AppserverID},
		)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListServerChannels(ctx, &channel.ListServerChannelsRequest{
			AppserverId: c0.AppserverID.String(),
		})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetChannels()))
	})

	t.Run("Success:can_filter_by_name", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		f := factory.NewFactory(ctx, db)
		s := f.Appserver(t, 0, nil)
		c0 := f.Channel(t, 0, &qx.Channel{Name: "foo", AppserverID: s.ID})
		f.Channel(t, 1, &qx.Channel{Name: "bar", AppserverID: s.ID})
		f.Channel(t, 2, nil)

		ctx = context.WithValue(
			ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: c0.AppserverID},
		)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListServerChannels(ctx, &channel.ListServerChannelsRequest{
			AppserverId: c0.AppserverID.String(), Name: wrapperspb.String("foo"),
		})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetChannels()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		serverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: mockAuth}

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
		mockQuerier.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})
}

func TestChannelRPCService_GetById(t *testing.T) {
	t.Run("Success:returns_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		c := f.Channel(t, 0, nil)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.GetById(
			ctx, &channel.GetByIdRequest{Id: c.ID.String(), AppserverId: su.Server.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, c.ID.String(), response.GetChannel().Id)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.GetById(
			ctx, &channel.GetByIdRequest{Id: uuid.NewString(), AppserverId: su.Server.ID.String()},
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
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.GetById(
			ctx,
			&channel.GetByIdRequest{Id: channelId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:invalid_uuid_returns_parsing_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
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
		ctx, db := testutil.Setup(t, func() {})
		mockId := uuid.NewString()

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &mockId, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

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
		mockAuth.AssertExpectations(t)
	})
}

func TestChannelRPCService_Create(t *testing.T) {
	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)

		svc := &rpcs.ChannelGRPCService{
			Deps: &rpcs.GrpcDependencies{Db: db, MProducer: testutil.MockRedisProducer}, Auth: testutil.TestMockAuth,
		}

		// ACT
		response, err := svc.Create(
			ctx,
			&channel.CreateRequest{Name: "new channel", AppserverId: su.Server.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.Channel)
	})

	t.Run("Error:when_create_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannel", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Create(
			ctx,
			&channel.CreateRequest{Name: "foo", AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

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
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionCreate).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

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
		mockAuth.AssertExpectations(t)
	})
}

func TestChannelRPCService_Delete(t *testing.T) {
	t.Run("Success:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		c := f.Channel(t, 0, nil)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
			ctx, &channel.DeleteRequest{Id: c.ID.String(), AppserverId: c.AppserverID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		c := f.Channel(t, 0, nil)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
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
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &mockId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

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
		mockAuth.AssertExpectations(t)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{ID: uuid.New()}, nil)
		mockQuerier.On("DeleteChannel", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.ChannelGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&channel.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})
}
