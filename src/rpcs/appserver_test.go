package rpcs_test

import (
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
	"mist/src/protos/v1/appserver"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverRPCService_List(t *testing.T) {
	t.Run("Successful:can_returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		// ACT
		response, err := testutil.TestAppserverClient.List(
			ctx, &appserver.ListRequest{Name: wrapperspb.String("random")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("Successful:can_return_all_resources_associated_with_user_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		testutil.TestAppserver(t, &qx.Appserver{Name: "foo", AppuserID: appuser.ID}, false)
		testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: appuser.ID}, false)

		// ACT
		response, err := testutil.TestAppserverClient.List(
			ctx, &appserver.ListRequest{},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

	t.Run("Successful:can_filter_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		aserver := testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: appuser.ID}, false)
		testutil.TestAppserver(t, nil, false)

		// ACT
		response, err := testutil.TestAppserverClient.List(
			ctx, &appserver.ListRequest{Name: wrapperspb.String(aserver.Name)},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetAppservers()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, mock.Anything, permission.ActionRead).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.List(ctx, &appserver.ListRequest{Name: &wrapperspb.StringValue{Value: "foo"}})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}

func TestAppserverRPCService_GetById(t *testing.T) {
	t.Run("Successful:returns_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aserver := testutil.TestAppserver(t, nil, false)

		// ACT
		response, err := testutil.TestAppserverClient.GetById(
			ctx, &appserver.GetByIdRequest{Id: aserver.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, aserver.ID.String(), response.GetAppserver().Id)
		assert.Equal(t, false, response.GetAppserver().IsOwner)
		assert.Equal(t, aserver.Name, response.GetAppserver().Name)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverClient.GetById(
			ctx, &appserver.GetByIdRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionRead).Return(nil)

		svc := &rpcs.AppserverGRPCService{
			Db: db.NewQuerier(qx.New(testutil.TestDbConn)), DbConn: testutil.TestDbConn, Auth: mockAuth,
		}

		// ACT
		response, err = svc.GetById(
			ctx, &appserver.GetByIdRequest{Id: uuid.NewString()},
		)
		s, ok = status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})

	t.Run("Error:invalid_uuid_returns_parsing_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

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
		ctx := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionRead).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.GetById(ctx, &appserver.GetByIdRequest{Id: "foo"})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}

func TestAppserverRPCService_Create(t *testing.T) {

	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		var count int
		ctx := testutil.Setup(t, func() {})
		userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)

		// ACT
		response, err := testutil.TestAppserverClient.Create(
			ctx, &appserver.CreateRequest{Name: "someone"},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		testutil.TestDbConn.QueryRow(ctx, "SELECT COUNT(*) FROM appserver").Scan(&count)

		serverSubs, _ := service.NewAppserverSubService(
			ctx, testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)), new(testutil.MockProducer),
		).ListUserServerSubs(appuser.ID)
		assert.NotNil(t, response.Appserver)
		assert.Equal(t, 1, len(serverSubs))
		assert.Equal(t, 1, count)
	})

	t.Run("Error:invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

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
		ctx := testutil.Setup(t, func() {})
		userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{AppuserID: userId, Name: "boo"}

		mockTxQuerier := new(testutil.MockQuerier)
		mockTxQuerier.On("CreateAppserver", ctx, expectedRequest).Return(nil, fmt.Errorf("a db error"))

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionCreate).Return(nil)

		svc := &rpcs.AppserverGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{
			Name: "boo",
		})

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "a db error")
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionCreate).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Create(ctx, &appserver.CreateRequest{
			Name: "boo",
		})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}

func TestAppserverRPCService_Delete(t *testing.T) {

	t.Run("Successful:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		aserver := testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: parsedUid}, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: aserver.ID, AppuserID: parsedUid}, false)

		subService := service.NewAppserverSubService(
			ctx, testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)), new(testutil.MockProducer),
		)

		// ASSERT
		serverSubs, _ := subService.ListUserServerSubs(appuser.ID)
		assert.Equal(t, 1, len(serverSubs))

		// ACT
		response, err := testutil.TestAppserverClient.Delete(
			ctx, &appserver.DeleteRequest{Id: aserver.ID.String()},
		)

		// ASSERT
		serverSubs, _ = subService.ListUserServerSubs(appuser.ID)
		assert.NotNil(t, response)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(serverSubs))
	})

	t.Run("Error:invalid_id_returns_unauthorized_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverClient.Delete(ctx, &appserver.DeleteRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.Contains(t, err.Error(), "(-5) Unauthorized")

	})

	t.Run("Error:on_database_failure_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, mock.Anything, permission.ActionDelete).Return(nil)
		// ACT
		svc := &rpcs.AppserverGRPCService{
			Db: db.NewQuerier(qx.New(testutil.TestDbConn)), DbConn: testutil.TestDbConn, Auth: mockAuth,
		}
		_, err := svc.Delete(ctx, &appserver.DeleteRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// // ASSERT
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())               // Check that the error code is NotFound
		assert.Contains(t, s.Message(), faults.NotFoundMessage) // Check the error message
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		roleId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, &roleId, permission.ActionDelete).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&appserver.DeleteRequest{Id: roleId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}
