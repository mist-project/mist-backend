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
	pb_appserverpermission "mist/src/protos/v1/appserver_permission"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverPermissionService_PgTypeToPb(t *testing.T) {
	// ARRANGE
	id := uuid.New()
	appserverId := uuid.New()
	appuserId := uuid.New()
	now := time.Now()

	perm := &qx.AppserverPermission{
		ID:          id,
		AppserverID: appserverId,
		AppuserID:   appuserId,
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}

	expected := &pb_appserverpermission.AppserverPermission{
		Id:          id.String(),
		AppserverId: appserverId.String(),
		AppuserId:   appuserId.String(),
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	svc := service.NewAppserverPermissionService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	// ACT
	res := svc.PgTypeToPb(perm)

	// ASSERT
	assert.Equal(t, expected, res)
}

func TestAppserverPermissionService_Create(t *testing.T) {
	t.Run("Successful:create_permission", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverPermissionParams{AppserverID: uuid.New(), AppuserID: uuid.New()}
		expected := qx.AppserverPermission{ID: uuid.New(), AppserverID: obj.AppserverID, AppuserID: obj.AppuserID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverPermission", ctx, obj).Return(expected, nil)

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		res, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverPermissionParams{AppserverID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverPermission", ctx, obj).Return(qx.AppserverPermission{}, fmt.Errorf("creation failed"))

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: creation failed")
	})
}

func TestAppserverPermissionService_ListAppserverPermissions(t *testing.T) {
	t.Run("Successful:list_permissions", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		expected := []qx.AppserverPermission{{ID: uuid.New(), AppserverID: appserverId, AppuserID: uuid.New()}}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverPermissions", ctx, appserverId).Return(expected, nil)

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		perms, err := svc.ListAppserverPermissions(appserverId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, perms)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverPermissions", ctx, appserverId).Return(nil, fmt.Errorf("db error"))

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.ListAppserverPermissions(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: db error")
	})
}

func TestAppserverPermissionService_GetById(t *testing.T) {
	t.Run("Successful:returns_permission_object", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()
		expected := qx.AppserverPermission{ID: id}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverPermissionById", ctx, id).Return(expected, nil)

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		actual, err := svc.GetById(id)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
	})

	t.Run("Error:returns_not_found_when_no_rows", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverPermissionById", ctx, id).Return(qx.AppserverPermission{}, fmt.Errorf(message.DbNotFound))

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetById(id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-2) resource not found")
	})

	t.Run("Error:returns_database_error_on_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverPermissionById", ctx, id).Return(qx.AppserverPermission{}, fmt.Errorf("boom"))

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetById(id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: boom")
	})
}

func TestAppserverPermissionService_Delete(t *testing.T) {
	t.Run("Successful:delete_permission", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverPermission", ctx, id).Return(int64(1), nil)

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(id)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverPermission", ctx, id).Return(int64(0), nil)

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:db_failure_on_delete", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverPermission", ctx, id).Return(nil, fmt.Errorf("db crash"))

		svc := service.NewAppserverPermissionService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) database error: db crash")
	})
}
