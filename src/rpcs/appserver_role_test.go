package rpcs_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestAppserveRoleService_ListServerRoles(t *testing.T) {
	t.Run("Successful:can_return_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleClient.ListServerRoles(
			ctx, &pb_appserverrole.ListServerRolesRequest{AppserverId: appserver.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppserverRoles()))
	})

	t.Run("Successful:can_return_all_appserver_roles_for_appserver_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil)
		testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "some random name", AppserverID: server.ID})
		testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "some random name #2", AppserverID: server.ID})

		// ACT
		response, err := testutil.TestAppserverRoleClient.ListServerRoles(
			ctx, &pb_appserverrole.ListServerRolesRequest{AppserverId: server.ID.String()},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppserverRoles()))
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.NewString()
		request := &pb_appserverrole.ListServerRolesRequest{
			AppserverId: appserverId,
		}
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverRoles", mock.Anything, mock.Anything).Return(
			[]qx.AppserverRole{}, fmt.Errorf("db error"),
		)

		svc := &rpcs.AppserverRoleGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn}

		// ACT
		response, err := svc.ListServerRoles(ctx, request)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "db error")
	})
}

func TestAppserveRoleService_Create(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleClient.Create(ctx, &pb_appserverrole.CreateRequest{
			AppserverId: appserver.ID.String(),
			Name:        "foo",
		})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.AppserverRole)
	})

	t.Run("Error:on_database_failure_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.Create(
			ctx, &pb_appserverrole.CreateRequest{Name: "foo", AppserverId: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
	})

	t.Run("Error:invalid_arguments_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.Create(ctx, &pb_appserverrole.CreateRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

func TestAppserveRoleService_Delete(t *testing.T) {
	t.Run("Successful:roles_can_only_be_deleted_by_server_owner_only", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		user := testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		server := testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: user.ID})
		aRole := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: server.ID, Name: "zoo"})

		// ACT
		response, err := testutil.TestAppserverRoleClient.Delete(
			ctx, &pb_appserverrole.DeleteRequest{Id: aRole.ID.String()},
		)

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error:cannot_be_deleted_by_non_owner", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		aRole := testutil.TestAppserverRole(t, nil)

		// ACT
		response, err := testutil.TestAppserverRoleClient.Delete(
			ctx, &pb_appserverrole.DeleteRequest{Id: aRole.ID.String()})

		// ASSERT
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-2) resource not found")
	})

	t.Run("Error:invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppserverRoleClient.Delete(ctx, &pb_appserverrole.DeleteRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})
}
