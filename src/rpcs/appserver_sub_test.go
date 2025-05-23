package rpcs_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mist/src/errors/message"
	"mist/src/permission"
	pb_appserver_sub "mist/src/protos/v1/appserver_sub"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestAppserverSubService_ListUserServerSubs(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.ListUserServerSubs(
			ctx, &pb_appserver_sub.ListUserServerSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("Successful:can_return_all_users_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		appserver := testutil.TestAppserver(t, nil, false)
		appserver2 := testutil.TestAppserver(t, nil, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver.ID, AppuserID: appuser.ID}, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: appserver2.ID, AppuserID: appuser.ID}, false)

		// ACT
		response, err := testutil.TestAppserverSubClient.ListUserServerSubs(
			ctx, &pb_appserver_sub.ListUserServerSubsRequest{},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

	t.Run("Error:on_authorization_error_it_errors", func(t *testing.T) {
		// ARRANGE
		var nilString *string
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", mock.Anything, nilString, permission.ActionRead, permission.SubActionListUserServerSubs).Return(
			message.UnauthorizedError("Unauthorized"),
		)

		svc := &rpcs.AppserverSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListUserServerSubs(
			ctx,
			&pb_appserver_sub.ListUserServerSubsRequest{},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}

func TestAppserverSubService_ListAppserverUserSubs(t *testing.T) {

	t.Run("Successful:can_return_all_appserver_subs_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user1 := testutil.TestAppuser(t, &qx.Appuser{ID: uuid.New(), Username: "foo"}, false)
		user2 := testutil.TestAppuser(t, &qx.Appuser{ID: uuid.New(), Username: "bar"}, false)
		sub := testutil.TestAppserverSub(t, nil, true)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: sub.AppserverID, AppuserID: user1.ID}, false)
		testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: sub.AppserverID, AppuserID: user2.ID}, false)

		// ACT
		response, err := testutil.TestAppserverSubClient.ListAppserverUserSubs(
			ctx,
			&pb_appserver_sub.ListAppserverUserSubsRequest{AppserverId: sub.AppserverID.String()},
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
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On(
			"Authorize", mock.Anything, nilString, permission.ActionRead, permission.SubActionListAppserverUserSubs,
		).Return(message.UnauthorizedError("Unauthorized"))

		svc := &rpcs.AppserverSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.ListAppserverUserSubs(
			ctx,
			&pb_appserver_sub.ListAppserverUserSubsRequest{AppserverId: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}

func TestAppserverSubService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: appuser.ID}, false)

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(
			ctx, &pb_appserver_sub.CreateRequest{AppserverId: appserver.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverSub)
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(
			ctx, &pb_appserver_sub.CreateRequest{AppserverId: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "database error")
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Create(
			ctx, &pb_appserver_sub.CreateRequest{},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestAppserverSubService_Delete(t *testing.T) {
	t.Run("Successful:deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		appuser := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"}, false)
		appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: appuser.ID}, false)
		appserverSub := testutil.TestAppserverSub(t, &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   appuser.ID,
		},
			false,
		)

		// ACT
		response, err := testutil.TestAppserverSubClient.Delete(
			ctx, &pb_appserver_sub.DeleteRequest{Id: appserverSub.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverSubClient.Delete(
			ctx, &pb_appserver_sub.DeleteRequest{Id: uuid.NewString()},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})

	t.Run("Error:when_db_fails_it_errors", func(t *testing.T) {
		// ARRANGE
		mockId := uuid.NewString()
		ctx := testutil.Setup(t, func() {})
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", ctx, mock.Anything).Return(nil, fmt.Errorf("db error"))
		mockAuth := new(testutil.MockAuthorizer)
		mockAuth.On("Authorize", ctx, &mockId, permission.ActionDelete, permission.SubActionDelete).Return(
			nil,
		)

		svc := &rpcs.AppserverSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_appserver_sub.DeleteRequest{Id: mockId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.Unknown, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
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

		svc := &rpcs.AppserverSubGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn, Auth: mockAuth}

		// ACT
		_, err := svc.Delete(
			ctx,
			&pb_appserver_sub.DeleteRequest{Id: roleId},
		)

		s, ok := status.FromError(err)

		// ASSERT
		assert.Equal(t, codes.PermissionDenied, s.Code())
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "(-5) Unauthorized")
	})
}
