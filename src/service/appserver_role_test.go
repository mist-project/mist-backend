package service_test

import (
	"context"
	"fmt"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserverrole "mist/src/protos/v1/appserver_role"
)

func TestAppserverRoleService_PgTypeToPb(t *testing.T) {
	// ARRANGE
	id := uuid.New()
	appserverID := uuid.New()
	now := time.Now()

	role := &qx.AppserverRole{
		ID:          id,
		AppserverID: appserverID,
		Name:        "admin",
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}

	expected := &pb_appserverrole.AppserverRole{
		Id:          id.String(),
		AppserverId: appserverID.String(),
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
		mockQuerier.On("CreateAppserverRole", ctx, obj).Return(qx.AppserverRole{}, fmt.Errorf("creation failed"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creation failed")
	})
}

func TestAppserverRoleService_ListAppserverRoles(t *testing.T) {
	t.Run("Successful:list_roles", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverID := uuid.New()
		expected := []qx.AppserverRole{{ID: uuid.New(), AppserverID: appserverID, Name: "admin"}}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverRoles", ctx, appserverID).Return(expected, nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		roles, err := svc.ListAppserverRoles(appserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, roles)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverRoles", ctx, appserverID).Return([]qx.AppserverRole{}, fmt.Errorf("db error"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.ListAppserverRoles(appserverID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestAppserverRoleService_DeleteByAppserver(t *testing.T) {
	t.Run("Successful:delete_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := qx.DeleteAppserverRoleParams{ID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", ctx, params).Return(int64(1), nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.DeleteByAppserver(params)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := qx.DeleteAppserverRoleParams{ID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", ctx, params).Return(int64(0), nil)

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.DeleteByAppserver(params)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:db_failure_on_delete", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := qx.DeleteAppserverRoleParams{ID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRole", ctx, params).Return(int64(0), fmt.Errorf("db crash"))

		svc := service.NewAppserverRoleService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.DeleteByAppserver(params)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db crash")
	})
}
