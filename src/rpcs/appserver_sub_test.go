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
	"mist/src/protos/v1/appserver_sub"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestAppserverSubRPCService_ListUserServerSubs(t *testing.T) {
	t.Run("Success:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListUserServerSubs(
			ctx, &appserver_sub.ListUserServerSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("Success:can_return_all_users_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		parsedUid, err := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		user := f.Appuser(t, 1, &qx.Appuser{ID: parsedUid, Username: "testuser"})
		s1 := f.Appserver(t, 1, nil)
		s2 := f.Appserver(t, 2, nil)

		f.AppserverSub(t, 1, &qx.AppserverSub{AppserverID: s1.ID, AppuserID: user.ID})
		f.AppserverSub(t, 2, &qx.AppserverSub{AppserverID: s2.ID, AppuserID: user.ID})

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListUserServerSubs(
			ctx, &appserver_sub.ListUserServerSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

	t.Run("Error:on_db_error_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)

		mockQuerier.On("ListUserServerSubs", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.ListUserServerSubs(
			ctx,
			&appserver_sub.ListUserServerSubsRequest{},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})
}

func TestAppserverSubRPCService_ListAppserverUserSubs(t *testing.T) {

	t.Run("Success:can_return_all_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverOwner(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		u0 := f.Appuser(t, 2, nil)
		u1 := f.Appuser(t, 3, nil)
		f.AppserverSub(t, 2, &qx.AppserverSub{AppserverID: su.Server.ID, AppuserID: u0.ID})
		f.AppserverSub(t, 3, &qx.AppserverSub{AppserverID: su.Server.ID, AppuserID: u1.ID})

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.ListAppserverUserSubs(
			ctx,
			&appserver_sub.ListAppserverUserSubsRequest{AppserverId: su.Server.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 3, len(response.GetAppusers()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nilString *string
		ctx, db := testutil.Setup(t, func() {})

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On(
			"Authorize", mock.Anything, nilString, permission.ActionRead,
		).Return(faults.AuthorizationError("Unauthorized", slog.LevelDebug))

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.ListAppserverUserSubs(
			ctx,
			&appserver_sub.ListAppserverUserSubsRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})
}

func TestAppserverSubRPCService_Create(t *testing.T) {
	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverUnsub(t, ctx, db)

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Create(ctx, &appserver_sub.CreateRequest{AppserverId: su.Server.ID.String()})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverSub)
	})

	t.Run("Error:invalid_db_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Create(
			ctx, &appserver_sub.CreateRequest{AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code())
		assert.Contains(t, s.Message(), faults.DatabaseErrorMessage)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(ctx, &appserver_sub.CreateRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestAppserverSubRPCService_Delete(t *testing.T) {
	t.Run("Success:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		tu := factory.UserAppserverSub(t, ctx, db)

		svc := &rpcs.AppserverSubGRPCService{
			Deps: &rpcs.GrpcDependencies{Db: db, MProducer: testutil.MockRedisProducer}, Auth: testutil.TestMockAuth,
		}

		// ACT
		response, err := svc.Delete(
			ctx, &appserver_sub.DeleteRequest{Id: tu.Sub.ID.String(), AppserverId: tu.Server.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: testutil.TestMockAuth}

		// ACT
		response, err := svc.Delete(
			ctx, &appserver_sub.DeleteRequest{Id: uuid.NewString(), AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, err.Error(), faults.NotFoundMessage)
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx, _ := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)

		mockQuerier.On("GetAppserverSubById", mock.Anything, mock.Anything).Return(qx.AppserverSub{}, nil)
		mockQuerier.On("DeleteAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}, Auth: testutil.TestMockAuth}

		// ACT
		_, err := svc.Delete(ctx, &appserver_sub.DeleteRequest{Id: mockId, AppserverId: uuid.NewString()})

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Internal, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		roleId := uuid.NewString()
		ctx, db := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			faults.AuthorizationError("Unauthorized", slog.LevelDebug),
		)

		svc := &rpcs.AppserverSubGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(ctx, &appserver_sub.DeleteRequest{Id: roleId, AppserverId: uuid.NewString()})

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), faults.AuthorizationErrorMessage)
		mockAuth.AssertExpectations(t)
	})
}
