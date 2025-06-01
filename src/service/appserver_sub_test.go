package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/appserver_sub"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverSubService_PgTypeToPb(t *testing.T) {

	// ARRANGE
	id := uuid.New()
	appserverId := uuid.New()
	now := time.Now()

	sub := &qx.AppserverSub{
		ID:          id,
		AppserverID: appserverId,
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}

	expected := &appserver_sub.AppserverSub{
		Id:          id.String(),
		AppserverId: appserverId.String(),
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	svc := service.NewAppserverSubService(
		context.Background(), testutil.TestDbConn, new(testutil.MockQuerier), new(testutil.MockProducer),
	)

	// ACT
	result := svc.PgTypeToPb(sub)

	// ASSERT
	assert.Equal(t, expected, result)
}

func TestAppserverSubService_PgAppserverSubRowToPb(t *testing.T) {

	// ARRANGE
	now := time.Now()
	row := &qx.ListUserServerSubsRow{
		ID:             uuid.New(),
		Name:           "Test Server",
		AppserverSubID: uuid.New(),
		CreatedAt:      pgtype.Timestamp{Time: now, Valid: true},
	}

	svc := service.NewAppserverSubService(
		context.Background(), testutil.TestDbConn, new(testutil.MockQuerier), new(testutil.MockProducer))

	// ACT
	pb := svc.PgAppserverSubRowToPb(row)

	// ASSERT
	assert.Equal(t, row.AppserverSubID.String(), pb.SubId)
	assert.Equal(t, row.ID.String(), pb.Appserver.Id)
	assert.Equal(t, row.Name, pb.Appserver.Name)
}

func TestAppserverSubService_PgUserSubRowToPb(t *testing.T) {

	// ARRANGE
	now := time.Now()
	row := &qx.ListAppserverUserSubsRow{
		ID:             uuid.New(),
		Username:       "tester",
		AppserverSubID: uuid.New(),
		CreatedAt:      pgtype.Timestamp{Time: now, Valid: true},
	}

	svc := service.NewAppserverSubService(
		context.Background(), testutil.TestDbConn, new(testutil.MockQuerier), new(testutil.MockProducer),
	)

	// ACT
	pb := svc.PgUserSubRowToPb(row)

	// ASSERT
	assert.Equal(t, row.ID.String(), pb.Appuser.Id)
	assert.Equal(t, row.Username, pb.Appuser.Username)
	assert.Equal(t, row.AppserverSubID.String(), pb.SubId)
}

func TestAppserverSubService_Create(t *testing.T) {

	t.Run("Successful:create_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverSubParams{AppserverID: uuid.New(), AppuserID: uuid.New()}
		expected := qx.AppserverSub{ID: uuid.New(), AppserverID: obj.AppserverID, AppuserID: obj.AppuserID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverSub", ctx, obj).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		result, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, result.ID)
	})

	t.Run("Error:failed_to_create", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverSubParams{AppserverID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverSub", ctx, obj).Return(nil, fmt.Errorf("create error"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)

		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: create error")
	})
}

func TestAppserverSubService_ListUserServerSubs(t *testing.T) {

	t.Run("Successful:list_subs_for_user", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		userID := uuid.New()
		expected := []qx.ListUserServerSubsRow{
			{
				ID:             uuid.New(),
				Name:           "Server1",
				AppserverSubID: uuid.New(),
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListUserServerSubs", ctx, userID).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		res, err := svc.ListUserServerSubs(userID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Error:on_db_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		userID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListUserServerSubs", ctx, userID).Return(
			[]qx.ListUserServerSubsRow{}, fmt.Errorf("db boom error"),
		)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.ListUserServerSubs(userID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db boom error")
	})
}

func TestAppserverSubService_ListAppserverUserSubs(t *testing.T) {

	t.Run("Successful:list_users_in_server", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		serverID := uuid.New()
		expected := []qx.ListAppserverUserSubsRow{
			{
				ID:             uuid.New(),
				Username:       "user1",
				AppserverSubID: uuid.New(),
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, serverID).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		res, err := svc.ListAppserverUserSubs(serverID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Error:on_db_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		serverID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, serverID).Return(nil, fmt.Errorf("query error"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.ListAppserverUserSubs(serverID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: query error")
	})
}

func TestAppserverSubService_GetById(t *testing.T) {

	t.Run("Successful:returns_appserver_sub_object", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expected := qx.AppserverSub{ID: uuid.New(), AppserverID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverSubById", ctx, expected.ID).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		actual, err := svc.GetById(expected.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.AppserverID, actual.AppserverID)
		assert.Equal(t, expected.AppuserID, actual.AppuserID)
	})

	t.Run("Error:returns_not_found_when_no_rows", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverSubById", ctx, appserverId).Return(nil, fmt.Errorf(message.DbNotFound))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find appserver sub with id: %v", appserverId))
	})

	t.Run("Error:returns_database_error_on_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverSubById", ctx, appserverId).Return(nil, fmt.Errorf("boom"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
	})
}

func TestAppserverSubService_Filter(t *testing.T) {

	t.Run("Successful:filters_subs", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		args := qx.FilterAppserverSubParams{
			AppuserID:   pgtype.UUID{Valid: true, Bytes: uuid.New()},
			AppserverID: pgtype.UUID{Valid: true, Bytes: uuid.New()},
		}
		expected := []qx.FilterAppserverSubRow{
			{
				ID:          uuid.New(),
				AppuserID:   uuid.New(),
				AppserverID: uuid.New(),
				CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("FilterAppserverSub", ctx, args).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		res, err := svc.Filter(args)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		args := qx.FilterAppserverSubParams{
			AppuserID:   pgtype.UUID{Valid: true, Bytes: uuid.New()},
			AppserverID: pgtype.UUID{Valid: true, Bytes: uuid.New()},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("FilterAppserverSub", ctx, args).Return(nil, fmt.Errorf("some db failure"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		res, err := svc.Filter(args)

		// ASSERT
		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: some db failure")
	})
}

func TestAppserverSubService_Delete(t *testing.T) {

	t.Run("Successful:deletes_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		mockSub := qx.AppserverSub{
			ID:        uuid.New(),
			AppuserID: uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", mock.Anything, mockSub.ID).Return(int64(1), nil)
		mockQuerier.On("GetAppserverSubById", mock.Anything, mockSub.ID).Return(mockSub, nil)

		mockProducer := new(testutil.MockProducer)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_SERVER, []*appuser.Appuser{
			{Id: mockSub.AppuserID.String()},
		}).Return(nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(mockSub.ID)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Error:returns_not_found_if_no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		mockSub := qx.AppserverSub{
			ID:        uuid.New(),
			AppuserID: uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", ctx, mockSub.ID).Return(int64(0), nil)
		mockQuerier.On("GetAppserverSubById", mock.Anything, mockSub.ID).Return(mockSub, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		err := svc.Delete(mockSub.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find appserver sub with id: %v", mockSub.ID))
	})

	t.Run("Error:returns_error_on_db_fail", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		mockSub := qx.AppserverSub{
			ID:        uuid.New(),
			AppuserID: uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", ctx, mockSub.ID).Return(int64(0), fmt.Errorf("db error"))
		mockQuerier.On("GetAppserverSubById", mock.Anything, mockSub.ID).Return(mockSub, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		err := svc.Delete(mockSub.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db error")
	})
}
