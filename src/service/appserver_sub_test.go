package service_test

import (
	"context"
	"fmt"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAppserverSubService_PgTypeToPb(t *testing.T) {
	id := uuid.New()
	appserverID := uuid.New()
	now := time.Now()

	sub := &qx.AppserverSub{
		ID:          id,
		AppserverID: appserverID,
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}

	expected := &pb_appserversub.AppserverSub{
		Id:          id.String(),
		AppserverId: appserverID.String(),
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	svc := service.NewAppserverSubService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	result := svc.PgTypeToPb(sub)

	assert.Equal(t, expected, result)
}

func TestAppserverSubService_PgAppserverSubRowToPb(t *testing.T) {
	now := time.Now()
	row := &qx.GetUserAppserverSubsRow{
		ID:             uuid.New(),
		Name:           "Test Server",
		AppserverSubID: uuid.New(),
		CreatedAt:      pgtype.Timestamp{Time: now, Valid: true},
	}

	svc := service.NewAppserverSubService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	pb := svc.PgAppserverSubRowToPb(row)

	assert.Equal(t, row.AppserverSubID.String(), pb.SubId)
	assert.Equal(t, row.ID.String(), pb.Appserver.Id)
	assert.Equal(t, row.Name, pb.Appserver.Name)
}

func TestAppserverSubService_PgUserSubRowToPb(t *testing.T) {
	now := time.Now()
	row := &qx.GetAllUsersAppserverSubsRow{
		ID:             uuid.New(),
		Username:       "tester",
		AppserverSubID: uuid.New(),
		CreatedAt:      pgtype.Timestamp{Time: now, Valid: true},
	}

	svc := service.NewAppserverSubService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	pb := svc.PgUserSubRowToPb(row)

	assert.Equal(t, row.ID.String(), pb.Appuser.Id)
	assert.Equal(t, row.Username, pb.Appuser.Username)
	assert.Equal(t, row.AppserverSubID.String(), pb.SubId)
}

func TestAppserverSubService_Create(t *testing.T) {
	t.Run("Successful:create_sub", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverSubParams{AppserverID: uuid.New(), AppuserID: uuid.New()}
		expected := qx.AppserverSub{ID: uuid.New(), AppserverID: obj.AppserverID, AppuserID: obj.AppuserID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverSub", ctx, obj).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		result, err := svc.Create(obj)

		assert.NoError(t, err)
		assert.Equal(t, expected.ID, result.ID)
	})

	t.Run("Error:failed_to_create", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateAppserverSubParams{AppserverID: uuid.New(), AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppserverSub", ctx, obj).Return(qx.AppserverSub{}, fmt.Errorf("create error"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		_, err := svc.Create(obj)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create error")
	})
}

func TestAppserverSubService_ListUserAppserverAndSub(t *testing.T) {
	t.Run("Successful:list_subs_for_user", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		userID := uuid.New()
		expected := []qx.GetUserAppserverSubsRow{
			{
				ID:             uuid.New(),
				Name:           "Server1",
				AppserverSubID: uuid.New(),
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetUserAppserverSubs", ctx, userID).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		res, err := svc.ListUserAppserverAndSub(userID)

		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Error:on_db_error", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		userID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetUserAppserverSubs", ctx, userID).Return(
			[]qx.GetUserAppserverSubsRow{}, fmt.Errorf("db boom error"),
		)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		_, err := svc.ListUserAppserverAndSub(userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db boom error")
	})
}

func TestAppserverSubService_ListAllUsersAppserverAndSub(t *testing.T) {
	t.Run("Successful:list_users_in_server", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		serverID := uuid.New()
		expected := []qx.GetAllUsersAppserverSubsRow{
			{
				ID:             uuid.New(),
				Username:       "user1",
				AppserverSubID: uuid.New(),
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAllUsersAppserverSubs", ctx, serverID).Return(expected, nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		res, err := svc.ListAllUsersAppserverAndSub(serverID)

		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Error:on_db_error", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		serverID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAllUsersAppserverSubs", ctx, serverID).Return([]qx.GetAllUsersAppserverSubsRow{}, fmt.Errorf("query error"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		_, err := svc.ListAllUsersAppserverAndSub(serverID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query error")
	})
}

func TestAppserverSubService_DeleteByAppserver(t *testing.T) {
	t.Run("Successful:deletes_sub", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		subID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", ctx, subID).Return(int64(1), nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		err := svc.DeleteByAppserver(subID)

		assert.NoError(t, err)
	})

	t.Run("Error:returns_not_found_if_no_rows_deleted", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		subID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", ctx, subID).Return(int64(0), nil)

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		err := svc.DeleteByAppserver(subID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:returns_error_on_db_fail", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		subID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserverSub", ctx, subID).Return(int64(0), fmt.Errorf("db error"))

		svc := service.NewAppserverSubService(ctx, testutil.TestDbConn, mockQuerier)

		err := svc.DeleteByAppserver(subID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
