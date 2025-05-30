package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	"mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverRoleService_PgTypeToPb(t *testing.T) {

	// ARRANGE
	id := uuid.New()
	appserverId := uuid.New()
	now := time.Now()

	role := &qx.AppserverRole{
		ID:          id,
		AppserverID: appserverId,
		Name:        "admin",
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}

	expected := &appserver_role.AppserverRole{
		Id:          id.String(),
		AppserverId: appserverId.String(),
		Name:        "admin",
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	svc := service.NewAppserverRoleService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	// ACT
	res := svc.PgTypeToPb(role)

	// ASSERT
	assert.Equal(t, expected, res)
}

func TestAppserverRoleService_Create(t *testing.T) {

	t.Run("Successful:create_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverRoleParams{AppserverID: uuid.New(), Name: "editor"}
		expected := qx.AppserverRole{ID: uuid.New(), AppserverID: obj.AppserverID, Name: obj.Name}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRole", ctx, obj).Return(expected, nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		res, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
		assert.Equal(t, obj.Name, res.Name)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverRoleParams{AppserverID: uuid.New(), Name: "viewer"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRole", ctx, obj).Return(nil, fmt.Errorf("creation failed"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: creation failed")
	})
}

func TestAppserverRoleService_ListAppserverRoles(t *testing.T) {

	t.Run("Successful:list_roles", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		expected := []qx.AppserverRole{{ID: uuid.New(), AppserverID: appserverId, Name: "admin"}}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverRoles", ctx, appserverId).Return(expected, nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		roles, err := svc.ListAppserverRoles(appserverId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, roles)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverRoles", ctx, appserverId).Return(nil, fmt.Errorf("db error"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.ListAppserverRoles(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
	})
}

func TestAppserverRoleService_GetAppuserRoles(t *testing.T) {

	t.Run("Successful:gets_roles", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expectedRequest := qx.GetAppuserRolesParams{
			AppserverID: uuid.New(),
			AppuserID:   uuid.New(),
		}
		expected := []qx.GetAppuserRolesRow{{ID: uuid.New(), Name: "admin"}}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppuserRoles", ctx, expectedRequest).Return(expected, nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		roles, err := svc.GetAppuserRoles(expectedRequest)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, roles)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expectedRequest := qx.GetAppuserRolesParams{
			AppserverID: uuid.New(),
			AppuserID:   uuid.New(),
		}
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppuserRoles", ctx, expectedRequest).Return(nil, fmt.Errorf("db error"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetAppuserRoles(expectedRequest)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
	})
}

func TestAppserverRoleService_GetById(t *testing.T) {

	t.Run("Successful:returns_appserver_role_object", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()
		expected := qx.AppserverRole{ID: roleId, Name: "test-app"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverRoleById", ctx, roleId).Return(expected, nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		actual, err := svc.GetById(roleId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
	})

	t.Run("Error:returns_not_found_when_no_rows", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverRoleById", ctx, appserverId).Return(nil, fmt.Errorf(message.DbNotFound))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-2) resource not found")
	})

	t.Run("Error:returns_database_error_on_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverRoleById", ctx, appserverId).Return(nil, fmt.Errorf("boom"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: boom")
	})
}

func TestAppserverRoleService_Delete(t *testing.T) {

	t.Run("Successful:delete_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", ctx, params).Return(int64(1), nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(params)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", ctx, params).Return(int64(0), nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(params)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:db_failure_on_delete", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", ctx, params).Return(nil, fmt.Errorf("db crash"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(params)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: db crash")
	})
}
