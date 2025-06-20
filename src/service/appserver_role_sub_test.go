package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/producer"
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

	svc := service.NewAppserverRoleSubService(
		context.Background(),
		&service.ServiceDeps{
			Db:        new(testutil.MockQuerier),
			MProducer: producer.NewMProducer(new(testutil.MockRedis)),
		},
	)

	// ACT
	res := svc.PgTypeToPb(roleSub)

	// ASSERT
	assert.Equal(t, roleSub.ID.String(), res.Id)
	assert.Equal(t, roleSub.AppserverRoleID.String(), res.AppserverRoleId)
	assert.Equal(t, roleSub.AppuserID.String(), res.AppuserId)
	assert.Equal(t, roleSub.AppserverID.String(), res.AppserverId)
}

func TestAppserverRoleSubService_Create(t *testing.T) {

	t.Run("Success:create_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
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
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateAppserverRoleSub", ctx, obj).Return(expected, nil)
		mockQuerier.On("GetChannelsForUsers", ctx, mock.Anything).Return([]qx.GetChannelsForUsersRow{}, nil)

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		res, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverRoleSubParams{
			AppserverRoleID: uuid.New(),
			AppuserID:       uuid.New(),
			AppserverID:     uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateAppserverRoleSub", ctx, obj).Return(nil, fmt.Errorf("insert failed"))

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: insert failed")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestAppserverRoleSubService_ListServerRoleSubs(t *testing.T) {

	t.Run("Success:fetch_role_subs", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		expected := []qx.ListServerRoleSubsRow{
			{AppuserID: uuid.New(), AppserverRoleID: uuid.New()},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListServerRoleSubs", ctx, appserverId).Return(expected, nil)

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		res, err := svc.ListServerRoleSubs(appserverId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListServerRoleSubs", ctx, appserverId).Return(
			[]qx.ListServerRoleSubsRow{}, fmt.Errorf("db fail"),
		)

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.ListServerRoleSubs(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db fail")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestAppserverRoleSubService_GetById(t *testing.T) {

	t.Run("Success:return_appserver_role_sub_object", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleId := uuid.New()
		expected := qx.AppserverRoleSub{
			ID: uuid.New(), AppuserID: uuid.New(), AppserverRoleID: uuid.New(), AppserverSubID: uuid.New(),
			AppserverID: uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverRoleSubById", ctx, roleId).Return(expected, nil)

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		actual, err := svc.GetById(roleId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.AppuserID, actual.AppuserID)
		assert.Equal(t, expected.AppserverRoleID, actual.AppserverRoleID)
		assert.Equal(t, expected.AppserverSubID, actual.AppserverSubID)
		assert.Equal(t, expected.AppserverID, actual.AppserverID)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:returns_not_found_when_no_rows", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverRoleSubById", ctx, appserverId).Return(nil, fmt.Errorf(message.DbNotFound))

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("no appserver role sub found for id: %s", appserverId))
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:returns_database_error_on_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverRoleSubById", ctx, appserverId).Return(nil, fmt.Errorf("boom"))

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestAppserverRoleSubService_Delete(t *testing.T) {

	t.Run("Success:delete_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleSub := qx.AppserverRoleSub{
			ID:              uuid.New(),
			AppserverRoleID: uuid.New(),
			AppuserID:       uuid.New(),
			AppserverSubID:  uuid.New(),
			AppserverID:     uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("DeleteAppserverRoleSub", ctx, roleSub.ID).Return(int64(1), nil)
		mockQuerier.On("GetAppserverRoleSubById", ctx, roleSub.ID).Return(roleSub, nil)
		mockQuerier.On("GetChannelsForUsers", ctx, mock.Anything).Return([]qx.GetChannelsForUsersRow{}, nil)

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(roleSub.ID)

		// ASSERT
		assert.NoError(t, err)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleSub := qx.AppserverRoleSub{
			ID:              uuid.New(),
			AppserverRoleID: uuid.New(),
			AppuserID:       uuid.New(),
			AppserverSubID:  uuid.New(),
			AppserverID:     uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("DeleteAppserverRoleSub", ctx, roleSub.ID).Return(int64(0), nil)
		mockQuerier.On("GetAppserverRoleSubById", ctx, roleSub.ID).Return(roleSub, nil)

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(roleSub.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(
			t, err, fmt.Sprintf("no appserver role sub found for id: %s", roleSub.ID),
		)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:db_failure_on_delete_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleSub := qx.AppserverRoleSub{
			ID:              uuid.New(),
			AppserverRoleID: uuid.New(),
			AppuserID:       uuid.New(),
			AppserverSubID:  uuid.New(),
			AppserverID:     uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverRoleSubById", ctx, roleSub.ID).Return(roleSub, nil)
		mockQuerier.On("DeleteAppserverRoleSub", ctx, roleSub.ID).Return(nil, fmt.Errorf("db crash"))

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(roleSub.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db crash")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:db_failure_on_get_role_sub_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleSub := qx.AppserverRoleSub{
			ID: uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverRoleSubById", ctx, roleSub.ID).Return(nil, fmt.Errorf("db crash"))

		svc := service.NewAppserverRoleSubService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(roleSub.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db crash")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}
