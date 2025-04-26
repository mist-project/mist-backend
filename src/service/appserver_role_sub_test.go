package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverRoleSubService_PgTypeToPb(t *testing.T) {
	// ARRANGE
	roleSub := &qx.AppserverRoleSub{
		ID:              uuid.New(),
		AppserverRoleID: uuid.New(),
		AppuserID:       uuid.New(),
		AppserverID:     uuid.New(),
	}

	svc := service.NewAppserverRoleSubService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	// ACT
	res := svc.PgTypeToPb(roleSub)

	// ASSERT
	assert.Equal(t, roleSub.ID.String(), res.Id)
	assert.Equal(t, roleSub.AppserverRoleID.String(), res.AppserverRoleId)
	assert.Equal(t, roleSub.AppuserID.String(), res.AppuserId)
	assert.Equal(t, roleSub.AppserverID.String(), res.AppserverId)
}

func TestAppserverRoleSubService_Create(t *testing.T) {
	t.Run("Successful:create_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverRoleSubParams{
			AppserverRoleID: uuid.New(),
			AppuserID:       uuid.New(),
			AppserverID:     uuid.New(),
		}
		expected := qx.AppserverRoleSub{
			ID:              uuid.New(),
			AppserverRoleID: obj.AppserverRoleID,
			AppuserID:       obj.AppuserID,
			AppserverID:     obj.AppserverID,
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRoleSub", ctx, obj).Return(expected, nil)

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		res, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverRoleSubParams{
			AppserverRoleID: uuid.New(),
			AppuserID:       uuid.New(),
			AppserverID:     uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverRoleSub", ctx, obj).Return(qx.AppserverRoleSub{}, fmt.Errorf("insert failed"))

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestAppserverRoleSubService_ListServerRoleSubs(t *testing.T) {
	t.Run("Successful:fetch_role_subs", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverID := uuid.New()
		expected := []qx.ListServerRoleSubsRow{
			{AppuserID: uuid.New(), AppserverRoleID: uuid.New()},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListServerRoleSubs", ctx, appserverID).Return(expected, nil)

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		res, err := svc.ListServerRoleSubs(appserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListServerRoleSubs", ctx, appserverID).Return(
			[]qx.ListServerRoleSubsRow{}, fmt.Errorf("db fail"),
		)

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.ListServerRoleSubs(appserverID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestAppserverRoleSubService_Delete(t *testing.T) {
	t.Run("Successful:delete_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.DeleteAppserverRoleSubParams{AppuserID: uuid.New(), ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRoleSub", ctx, obj).Return(int64(1), nil)

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(obj)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.DeleteAppserverRoleSubParams{AppuserID: uuid.New(), ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRoleSub", ctx, obj).Return(int64(0), nil)

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.DeleteAppserverRoleSubParams{AppuserID: uuid.New(), ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverRoleSub", ctx, obj).Return(int64(0), fmt.Errorf("db crash"))

		svc := service.NewAppserverRoleSubService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}
