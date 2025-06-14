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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"mist/src/faults"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/service"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestAppserverRPCService_List(t *testing.T) {
	t.Run("Success:can_returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.List(
			ctx, &appserver.ListRequest{Name: wrapperspb.String("random")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("Success:can_return_all_resources_associated_with_user_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		factory.UserAppserverOwner(t, ctx, db)
		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.List(
			ctx, &appserver.ListRequest{},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetAppservers()))
	})

	t.Run("Success:can_filter_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.List(
			ctx, &appserver.ListRequest{Name: wrapperspb.String(su.Server.Name)},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetAppservers()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)

		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.List(ctx, &appserver.ListRequest{Name: &wrapperspb.StringValue{Value: "foo"}})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})
}

func TestAppserverRPCService_GetById(t *testing.T) {
	t.Run("Success:returns_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.GetById(
			ctx, &appserver.GetByIdRequest{Id: server.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, server.ID.String(), response.GetAppserver().Id)
		assert.Equal(t, false, response.GetAppserver().IsOwner)
		assert.Equal(t, server.Name, response.GetAppserver().Name)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.GetById(
			ctx, &appserver.GetByIdRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), faults.NotFoundMessage)
	})

	t.Run("Error:invalid_uuid_returns_parsing_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverClient.GetById(
			ctx, &appserver.GetByIdRequest{Id: "foo"},
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

		mockAuth := new(testutil.MockAuthorizer)

		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionRead).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.GetById(ctx, &appserver.GetByIdRequest{Id: "foo"})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})
}

func TestAppserverRPCService_Create(t *testing.T) {

	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		claims, _ := middleware.GetJWTClaims(ctx)
		serverId := uuid.New()
		expectedRequest := qx.CreateAppserverParams{AppuserID: uuid.MustParse(claims.UserID), Name: "boo"}
		expectedSub := qx.CreateAppserverSubParams{AppuserID: uuid.MustParse(claims.UserID), AppserverID: serverId}

		mockQuerier := new(testutil.MockQuerier)
		mockTxQuerier := new(testutil.MockQuerier)

		mockQuerier.On("Begin", mock.Anything).Return(mockTxQuerier, nil)
		mockTxQuerier.On("CreateAppserver", ctx, expectedRequest).Return(qx.Appserver{ID: serverId}, nil)
		mockTxQuerier.On("CreateAppserverSub", ctx, expectedSub).Return(qx.AppserverSub{}, nil)
		mockTxQuerier.On("Commit", ctx).Return(nil)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}
		// ACT
		_, err := svc.Create(
			ctx, &appserver.CreateRequest{Name: expectedRequest.Name},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		mockTxQuerier.AssertExpectations(t)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverClient.Create(ctx, &appserver.CreateRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error:\n - name: value length must be at least 1 characters")
	})

	t.Run("Error:error_on_db_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{AppuserID: userId, Name: "boo"}

		mockQuerier := new(testutil.MockQuerier)
		mockTxQuerier := new(testutil.MockQuerier)

		mockQuerier.On("Begin", mock.Anything).Return(mockTxQuerier, nil)
		mockTxQuerier.On("CreateAppserver", ctx, expectedRequest).Return(nil, fmt.Errorf("a db error"))
		mockTxQuerier.On("Rollback", ctx).Return(nil)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{Name: "boo"})

		// ASSERT
		s, ok := status.FromError(err)
		assert.NotNil(t, err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code()) // Check that the error code is Internal
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockTxQuerier.AssertExpectations(t)
	})

	t.Run("Error:error_on_db_begin_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("Begin", mock.Anything).Return(nil, faults.DatabaseError("a db error", slog.LevelError))

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{Name: "boo"})

		// ASSERT
		s, ok := status.FromError(err)
		assert.NotNil(t, err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code()) // Check that the error code is Internal
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:error_on_commit_rollback_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{AppuserID: userId, Name: "boo"}

		mockQuerier := new(testutil.MockQuerier)
		mockTxQuerier := new(testutil.MockQuerier)

		mockQuerier.On("Begin", mock.Anything).Return(mockTxQuerier, nil)
		mockTxQuerier.On("CreateAppserver", ctx, expectedRequest).Return(qx.Appserver{ID: uuid.New()}, nil)
		mockTxQuerier.On("CreateAppserverSub", ctx, mock.Anything).Return(qx.AppserverSub{ID: uuid.New()}, nil)
		mockTxQuerier.On("Commit", ctx, mock.Anything).Return(fmt.Errorf("a db error"))

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{Name: "boo"})

		// ASSERT
		s, ok := status.FromError(err)
		assert.NotNil(t, err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code()) // Check that the error code is Internal
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:error_on_db_and_rollback_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{AppuserID: userId, Name: "boo"}

		mockQuerier := new(testutil.MockQuerier)
		mockTxQuerier := new(testutil.MockQuerier)

		mockQuerier.On("Begin", mock.Anything).Return(mockTxQuerier, nil)
		mockTxQuerier.On("CreateAppserver", ctx, expectedRequest).Return(nil, fmt.Errorf("a db error"))
		mockTxQuerier.On("Rollback", ctx).Return(fmt.Errorf("boom"))

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{Name: "boo"})

		// ASSERT
		s, ok := status.FromError(err)
		assert.NotNil(t, err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code()) // Check that the error code is Internal
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
		mockTxQuerier.AssertExpectations(t)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionCreate).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{
			Name: "boo",
		})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})
}

func TestAppserverRPCService_Delete(t *testing.T) {

	t.Run("Success:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		sub := factory.NewFactory(ctx, db).AppserverSub(t, 0, nil)

		subService := service.NewAppserverSubService(
			ctx,
			&service.ServiceDeps{
				Db:        db,
				MProducer: testutil.MockRedisProducer,
			},
		)

		// ASSERT
		serverSubs, _ := subService.ListUserServerSubs(sub.AppuserID)
		assert.Equal(t, 1, len(serverSubs))

		svc := &rpcs.AppserverGRPCService{
			Deps: &rpcs.GrpcDependencies{Db: db, MProducer: testutil.MockRedisProducer}, Auth: testutil.TestMockAuth,
		}

		// ACT
		response, err := svc.Delete(
			ctx, &appserver.DeleteRequest{Id: sub.AppserverID.String()},
		)

		// ASSERT
		serverSubs, _ = subService.ListUserServerSubs(sub.AppuserID)
		assert.NotNil(t, response)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(serverSubs))
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(ctx, &appserver.DeleteRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())

	})

	t.Run("Error:on_database_failure_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, mock.Anything).Return(nil, fmt.Errorf("a db error"))

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Delete(ctx, &appserver.DeleteRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// // ASSERT
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code())                    // Check that the error code is NotFound
		assert.Contains(t, s.Message(), faults.DatabaseErrorMessage) // Check the error message
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		roleId := uuid.NewString()
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)

		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&appserver.DeleteRequest{Id: roleId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})
}
